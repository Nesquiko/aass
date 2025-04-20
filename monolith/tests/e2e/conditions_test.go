//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/server"
)

func TestGetConditionById(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.getcond.%s@example.com", uuid.NewString())
	patientReq := newPatient(patientEmail)
	createdPatient := mustCreatePatient(t, patientReq)
	require.NotEmpty(t, createdPatient.Id, "Setup failed: Created patient ID is empty")

	startTime := time.Now().Truncate(time.Second)
	conditionName := "Hypertension"
	newCondition := api.NewCondition{
		Name:      conditionName,
		PatientId: createdPatient.Id,
		Start:     startTime,
	}

	reqBody, err := json.Marshal(newCondition)
	require.NoError(t, err, "Failed to marshal NewCondition request")

	createURL := fmt.Sprintf("%s/conditions", ServerUrl)
	createResp, err := http.Post(createURL, server.ApplicationJSON, bytes.NewReader(reqBody))
	require.NoError(t, err, "http.Post failed for CreatePatientCondition")
	defer createResp.Body.Close()

	bodyBytes, err := io.ReadAll(createResp.Body)
	require.NoError(t, err, "Failed to read CreatePatientCondition response body")
	createResp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(
		t,
		http.StatusCreated,
		createResp.StatusCode,
		"Expected 201 Created. Body: %s",
		string(bodyBytes),
	)

	var createdCond api.ConditionDisplay
	err = json.NewDecoder(createResp.Body).Decode(&createdCond)
	require.NoError(t, err, "Failed to decode CreatePatientCondition response")
	require.NotNil(t, createdCond.Id, "Created condition ID should not be nil")
	require.NotEmpty(t, *createdCond.Id, "Created condition ID should not be empty")

	getURL := fmt.Sprintf("%s/conditions/%s", ServerUrl, *createdCond.Id)
	getResp, err := http.Get(getURL)
	require.NoError(t, err, "http.Get failed for ConditionDetail")
	defer getResp.Body.Close()

	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "Failed to read ConditionDetail response body")
	getResp.Body = io.NopCloser(bytes.NewBuffer(getBody))

	require.Equal(
		t,
		http.StatusOK,
		getResp.StatusCode,
		"Expected 200 OK. Body: %s",
		string(getBody),
	)

	var fetchedCond api.ConditionDisplay
	err = json.NewDecoder(getResp.Body).Decode(&fetchedCond)
	require.NoError(t, err, "Failed to decode ConditionDetail response")

	assert := assert.New(t)
	assert.Equal(*createdCond.Id, *fetchedCond.Id, "Condition ID mismatch")
	assert.Equal(createdCond.Name, fetchedCond.Name, "Condition name mismatch")
	assert.WithinDuration(
		createdCond.Start,
		fetchedCond.Start,
		time.Second,
		"Condition start time mismatch",
	)
	assert.Equal(createdCond.End, fetchedCond.End, "Condition end mismatch")
}

func TestCreatePatientCondition(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.patient.cond.ok.%s@example.com", uuid.NewString())
	patientRequest := newPatient(patientEmail)
	createdPatient := mustCreatePatient(t, patientRequest)
	require.NotEmpty(t, createdPatient.Id, "Setup failed: Created patient ID is empty")

	startTime := time.Now().Truncate(time.Second)
	conditionName := "Acute Bronchitis"
	newConditionRequest := api.NewCondition{
		Name:      conditionName,
		PatientId: createdPatient.Id,
		Start:     startTime,
	}

	reqBodyBytes, err := json.Marshal(newConditionRequest)
	require.NoError(t, err, "Failed to marshal NewCondition request")

	url := fmt.Sprintf("%s/conditions", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(reqBodyBytes))
	require.NoError(t, err, "http.Post failed for CreatePatientCondition")
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	require.NoError(t, readErr, "Failed to read response body")
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	require.Equal(
		t,
		http.StatusCreated,
		res.StatusCode,
		"Expected '201 Created' status code. Response body: %s",
		string(bodyBytes),
	)

	var createdCondition api.ConditionDisplay
	err = json.NewDecoder(res.Body).Decode(&createdCondition)
	require.NoError(t, err, "Failed to decode successful response body into ConditionDisplay")

	assert := assert.New(t)
	assert.NotNil(createdCondition.Id, "Response condition ID should not be nil")
	assert.NotEmpty(*createdCondition.Id, "Response condition ID should not be empty")
	assert.Equal(conditionName, createdCondition.Name, "Response condition name mismatch")
	assert.WithinDuration(
		startTime,
		createdCondition.Start,
		time.Second,
		"Response condition start time mismatch",
	)
	assert.Nil(
		createdCondition.End,
		"Response condition end time should be nil as it wasn't provided",
	)
}

func mustCreateCondition(t *testing.T, request api.NewCondition) api.ConditionDisplay {
	t.Helper()
	require := require.New(t)

	reqBodyBytes, err := json.Marshal(request)
	require.NoError(err, "mustCreateCondition: Failed to marshal request")

	url := fmt.Sprintf("%s/conditions", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(reqBodyBytes))
	require.NoError(err, "mustCreateCondition: http.Post failed")
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	require.NoError(readErr, "mustCreateCondition: Failed to read response body")
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(
		http.StatusCreated,
		res.StatusCode,
		"mustCreateCondition: Expected '201 Created'. Body: %s",
		string(bodyBytes),
	)

	var createdCondition api.ConditionDisplay
	err = json.NewDecoder(res.Body).Decode(&createdCondition)
	require.NoError(err, "mustCreateCondition: Failed to decode response")
	require.NotNil(createdCondition.Id, "mustCreateCondition: Response ID is nil")
	require.NotEmpty(*createdCondition.Id, "mustCreateCondition: Response ID is empty UUID")

	return createdCondition
}
