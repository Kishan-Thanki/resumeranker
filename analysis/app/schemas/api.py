from pydantic import BaseModel

class EngineRequest(BaseModel):
    resume_pdf: str 
    job_description_pdf: str 

class EngineResponse(BaseModel):
    result_json: str
    model: str
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
