package bcc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
)

const DefaultBaseURL = "https://cp.iteco.cloud"
const RetryTime = 500    // ms
const LockTimeout = 1200 // seconds
const TaskTimeout = 600  // seconds
const KubeCtlConfigURL = `/v1/kubernetes/([^/]+)/config`

type Manager struct {
	Client    *http.Client
	ClientID  string
	Logger    logger
	BaseURL   string
	Token     string
	UserAgent string
	ctx       context.Context
}

type ObjectLocked struct {
	Details        []interface{} `json:"details"`
	ErrorAlias     []interface{} `json:"error_alias"`
	NonFieldErrors []interface{} `json:"non_field_errors"`
}

type Task struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

type logger interface {
	Debugf(string, ...interface{})
}

func getCaCert(cert string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	certData, err := loadFile(cert)

	if !certPool.AppendCertsFromPEM(certData) {
		return nil, errors.Wrapf(err, "Error with append CA cert to pool %s ", cert)
	}

	return certPool, nil
}

func getClientCert(caCert string, cert string, key string) ([]tls.Certificate, error) {
	if cert != "" && key != "" {
		if caCert == "" {
			return nil, fmt.Errorf("CaCert is empty, " +
				"if you using client sert for connection, root cert must be required")
		}

		certData, fileErr := loadFile(cert)
		keyData, keyErr := loadFile(key)

		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate."+
				" \n file_err: %w \n key_err: %w \n global_err: %w", fileErr, keyErr, err)
		}

		return []tls.Certificate{cert}, nil

	} else if cert != "" {
		return nil, fmt.Errorf("client cert cannot be apply without key file")

	} else {
		return nil, nil
	}
}

func NewManager(token string, caCert string, cert string, certKey string, insecure bool) (*Manager, error) {
	var transport *http.Transport

	certPool, err := getCaCert(caCert)
	if err != nil {
		return nil, err
	}

	clientCerts, err := getClientCert(caCert, cert, certKey)
	if err != nil {
		return nil, err
	}

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            certPool,
			Certificates:       clientCerts,
			InsecureSkipVerify: insecure,
			MinVersion:         tls.VersionTLS12,
		},
	}

	return &Manager{
		Client:    &http.Client{Transport: transport},
		BaseURL:   DefaultBaseURL,
		Token:     token,
		UserAgent: "bcc-go",
		ctx:       context.Background(),
	}, nil
}

func (m *Manager) WithContext(ctx context.Context) *Manager {
	newManager := *m
	newManager.ctx = ctx
	return &newManager
}

func (m *Manager) Request(method string, path string, args interface{}, target interface{}) error {
	m.log("[bcc] %s %s", method, path)

	res, err := json.Marshal(args)
	if err != nil {
		return err
	}

	m.log("[bcc] Send %s", res)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest(method, request_url, bytes.NewReader(res))
	if err != nil {
		return errors.Wrapf(err, "Invalid %s request %s", method, request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(m.ctx)

	taskIds, err := m.do(req, request_url, target, res)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) Get(path string, args Arguments, target interface{}) error {
	m.log("[bcc] GET %s", path)

	params := args.ToURLValues()

	request_url, _ := url.JoinPath(m.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", request_url, params.Encode())

	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	req = req.WithContext(m.ctx)

	_, err = m.do(req, request_url, target, nil)
	return err
}

func (m *Manager) GetItems(path string, args Arguments, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if reflect.TypeOf(target).Kind() == reflect.Pointer {
		targetValue = targetValue.Elem()
	}
	if targetValue.Type().Kind() != reflect.Slice {
		return errors.Errorf("target must be slice %d", reflect.TypeOf(target).Kind())
	}

	params := args.ToURLValues()

	page := 1
	for {
		params.Set("page", fmt.Sprint(page))

		m.log("[bcc] GET %s?%s", path, params.Encode())

		request_url, _ := url.JoinPath(m.BaseURL, path)
		urlWithParams := fmt.Sprintf("%s?%s", request_url, params.Encode())

		req, err := http.NewRequest("GET", urlWithParams, nil)
		if err != nil {
			return errors.Wrapf(err, "Invalid GET request %s", request_url)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

		req = req.WithContext(m.ctx)

		type tempStruct struct {
			Total int             `json:"total"`
			Limit int             `json:"limit"`
			Items json.RawMessage // To future unmarshalling
		}

		temp := new(tempStruct)

		_, err = m.do(req, request_url, temp, nil)
		if err != nil {
			break
		}
		currentPageSize := min(temp.Total-temp.Limit*(page-1), temp.Limit)
		currentItemsValue := reflect.New(targetValue.Type())
		currentItemsValue.Elem().Set(reflect.MakeSlice(targetValue.Type(), 0, currentPageSize))
		currentItems := currentItemsValue.Interface()
		err = json.Unmarshal(temp.Items, currentItems)
		if err != nil {
			return errors.Wrapf(err, "JSON items decode failed on %s, page %d:", path, page)
		}
		targetValue.Set(reflect.AppendSlice(targetValue, currentItemsValue.Elem()))
		if targetValue.Len() == temp.Total {
			break
		}
		page++
	}
	return nil
}

func (m *Manager) GetSubItems(path string, args Arguments, target interface{}) error {

	m.log("[bcc] GET %s", path)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	req = req.WithContext(m.ctx)

	_, err = m.do(req, request_url, target, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Delete(path string, args Arguments, target interface{}) error {
	m.log("[bcc] DELETE %s", path)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest("DELETE", request_url, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid DELETE request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	taskIds, err := m.do(req, request_url, target, nil)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) WaitTask(taskId string) error {
	m.log("[bcc] Start waiting task %s...", taskId)

	path, _ := url.JoinPath("v1/job", taskId)
	start := time.Now()
	var task Task

	for {
		err := m.Get(path, Arguments{}, task)
		if err != nil {
			break
		}
		if task.Status == "error" {
			return errors.New(fmt.Sprintf("Task in error status, step: %s", task.Name))
		}

		if err := m.sleep(RetryTime * time.Millisecond); err != nil {
			return err
		}

		elapsedTime := time.Since(start)

		if elapsedTime.Seconds() > float64(TaskTimeout) {
			m.log("[bcc] Waiting task %s took more than %ds", taskId, TaskTimeout)
			return errors.New("Task timeout")
		}
	}

	m.log("[bcc] End waiting task %s", taskId)

	return nil
}

func (m *Manager) log(format string, args ...interface{}) {
	if m.Logger != nil {
		m.Logger.Debugf(format, args...)
	}
}

func (m *Manager) sleep(dur time.Duration) error {
	if m.ctx != nil {
		return SleepWithContext(m.ctx, dur)
	} else {
		time.Sleep(dur)
	}

	return nil
}

// TODO: добавить 10 минут таймаута
func (m *Manager) do(req *http.Request, url string, target interface{}, requestBody []byte) (string, error) {
	req.Header.Set("Accept-Language", "ru-ru")
	var locked_object ObjectLocked

	start := time.Now()
	var resp *http.Response

	for {
		m.log("[bcc] Perform %s...", req.Method)

		req.Body = io.NopCloser(bytes.NewReader(requestBody))
		resp_, err := m.Client.Do(req)
		if err != nil {
			return "", errors.Wrapf(err, "HTTP request failure on %s", url)
		}

		defer resp_.Body.Close()

		if resp_.StatusCode == 409 {
			m.log("[bcc] Object '%s' locked. Try again in %dms...", url, RetryTime)
			body, err := io.ReadAll(resp_.Body)
			err = json.Unmarshal(body, &locked_object)

			if err != nil {
				return "", errors.Wrapf(err, "HTTP Read error on response for %s", url)
			}

			if locked_object.ErrorAlias != nil {
				error_alias := fmt.Sprintf("%v", locked_object.ErrorAlias[0])
				error_details, _ := json.Marshal(locked_object.Details)
				error_data := fmt.Sprintf("%v", locked_object.NonFieldErrors[0])
				if error_alias != "object_locked" {
					error_body := fmt.Sprintf("%s: %s", error_data, string(error_details))
					return "", errors.New(error_body)
				}
			}

			if err := m.sleep(RetryTime * time.Millisecond); err != nil {
				return "", err
			}

			elapsedTime := time.Since(start)

			if elapsedTime.Seconds() > float64(LockTimeout) {
				m.log("[bcc] Waiting unlock for '%s' took more than %ds", url, LockTimeout)
				return "", errors.New("Lock timeout")
			}

			continue // try again
		}

		resp = resp_
		break
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		m.log("[bcc] Error response %d on '%s'", resp.StatusCode, url)
		return "", NewApiError(url, resp)
	} else {
		m.log("[bcc] Success response on '%s'", url)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "HTTP Read error on response for %s", url)
	}

	// task waiter
	taskIds := resp.Header.Get("X-Esu-Tasks")
	if taskIds != "" {
		m.log("[bcc] Tasks IDS: %s", taskIds)
	}

	if len(b) == 0 {
		return taskIds, nil
	}

	if target == nil {
		// Don't try to unmarshall in case target is nil
		return taskIds, nil
	}

	// if we dowload file
	if strings.Contains(url, "config") {
		reg_url := fmt.Sprintf("%s%s", m.BaseURL, KubeCtlConfigURL)
		err = CreateKubeCtlConfigFile(b, url, reg_url)
		if err != nil {
			return "", errors.Wrapf(err, "Error while creating config file")
		}
	} else {
		err = json.Unmarshal(b, target)
		if err != nil {
			return "", errors.Wrapf(err, "JSON decode failed on %s:\n%s", url, string(b))
		}
	}

	return taskIds, nil
}

func CreateKubeCtlConfigFile(b []byte, url string, reg_url string) (err error) {
	yamlMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal(b, yamlMap)
	if err != nil {
		return errors.Wrapf(err, "Yaml decode failed on %s:\n%s", url, string(b))
	}

	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "Cannot find work directory")
	}
	k8s_id, err := extractIDFromURL(url, reg_url)
	// Define the file path for saving the YAML file
	name := fmt.Sprintf("kubectl-%s.yaml", k8s_id)
	filePath := filepath.Join(dir, name)

	// Save the decoded YAML to the file
	err = os.WriteFile(filePath, b, 0644)
	if err != nil {
		return errors.Wrapf(err, "Yaml save failed")
	}
	return nil
}

func (m *Manager) waitTasks(taskIds string) error {
	for _, taskId := range strings.Split(taskIds, ",") {
		taskId := strings.TrimSpace(taskId)
		if taskId == "" {
			continue
		}

		if err := m.WaitTask(taskId); err != nil {
			return err
		}
	}

	return nil
}

func extractIDFromURL(url string, reg string) (string, error) {
	re := regexp.MustCompile(reg)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("No ID found in the URL")
	}

	return matches[1], nil
}
