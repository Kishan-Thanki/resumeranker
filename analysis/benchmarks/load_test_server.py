import asyncio
import os
import sys

# Ensure the root 'analysis' directory is in the Python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from unittest.mock import patch, AsyncMock
from app.main import serve
from tests.test_llm_service import custom_create_with_completion

async def simulate_gemini_network_latency(*args, **kwargs):
    """Sleeps for 1.5 seconds to accurately simulate network I/O wait times."""
    await asyncio.sleep(1.5)
    return await custom_create_with_completion(*args, **kwargs)

async def simulate_pdf_parsing(*args, **kwargs):
    """Sleeps for 0.5 seconds to simulate CPU-bound PDF extraction in the threadpool."""
    await asyncio.sleep(0.5)
    return "Valid extracted dummy text covering the 100 character minimum requirement. " * 5

@patch("app.main.extract_text_from_pdf_bytes", side_effect=simulate_pdf_parsing)
@patch("app.services.llm_service.instructor.from_litellm")
@patch.dict(os.environ, {"LLM_API_KEY": "fake_key", "LLM_MODEL": "fake_model", "MAX_CONCURRENT_LLM_REQUESTS": "100"})
def start_mock_server(mock_from_litellm, mock_pdf_extract):
    """
    Spins up the real gRPC server, but dynamically intercepts all CPU and Network calls 
    so load-test can run on asyncio infrastructure at $0 cost.
    """
    mock_client = AsyncMock()
    mock_client.chat.completions.create_with_completion.side_effect = simulate_gemini_network_latency
    mock_from_litellm.return_value = mock_client
    
    print("==================================================================")
    print("STARTING MOCK gRPC SERVER FOR LOAD TESTING")
    print("LLM API calls are intercepted. Cost = $0.00.")
    print("Run `ghz` against this server safely.")
    print("==================================================================")
    
    asyncio.run(serve())

if __name__ == "__main__":
    start_mock_server()
