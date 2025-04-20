//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/Nesquiko/wac/pkg/api"
)

func TestRestartPersistence(t *testing.T) {
	patientEmail := fmt.Sprintf("test.patient.persistence.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	restartServer(t)

	url := fmt.Sprintf("%s/patients/%s", ServerUrl, patient.Id)
	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for patient")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedPatient api.Patient
	err = json.NewDecoder(res.Body).Decode(&fetchedPatient)
	require.NoError(t, err, "Failed to decode fetched patient")

	assert := assert.New(t)
	assert.Equal(patient.Id, fetchedPatient.Id, "Patient ID mismatch")
	assert.Equal(patient.Email, fetchedPatient.Email, "Patient email mismatch")
}
