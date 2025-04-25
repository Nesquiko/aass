package pkg

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Nesquiko/aass/medical-history-service/api"
)

// PatientsMedicalHistoryFiles implements api.ServerInterface.
func (s MedicalHistoryServer) PatientsMedicalHistoryFiles(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.PatientsMedicalHistoryFilesParams,
) {
	files := make([]string, 0)
	for i := range params.PageSize {
		fileType := dummyFileTypes[rand.Intn(len(dummyFileTypes))]
		date := time.Now().AddDate(0, 0, -rand.Intn(365)).Format("2006-01-02")
		files = append(files, fmt.Sprintf("%s_%s_%d.pdf", fileType, date, i))
	}

	pagination := api.Pagination{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    38,
	}

	encode(w, http.StatusOK, api.MedicalHistoryFileList{Files: files, Pagination: pagination})
}

var dummyFileTypes = []string{
	"lab_result",
	"medical_report",
	"prescription",
	"discharge_summary",
	"consultation_note",
	"radiology_report",
	"pathology_report",
	"medical_certificate",
	"referral_letter",
	"follow_up_note",
	"progress_note",
	"medication_list",
	"allergy_report",
	"immunization_record",
	"vital_signs_chart",
	"medical_history",
	"family_medical_history",
	"surgical_report",
	"anesthesia_report",
	"recovery_room_note",
	"physical_examination_report",
	"diagnostic_test_result",
	"treatment_plan",
	"care_plan",
	"discharge_plan",
}
