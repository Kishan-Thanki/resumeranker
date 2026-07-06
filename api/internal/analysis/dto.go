package analysis

type ProcessResumeRequest struct {
	ResumeText     string `json:"resume_text"`
	JobDescription string `json:"job_description"`
}
