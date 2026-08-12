#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "Generating Python gRPC code..."
uv run python -m grpc_tools.protoc \
  -I ../proto \
  --python_out=app/pb \
  --grpc_python_out=app/pb \
  --mypy_out=app/pb \
  --mypy_grpc_out=app/pb \
  ../proto/analysis.proto

# Fix the sibling-module import so it works when analysis_pb2_grpc.py is
# imported as part of the app.pb package rather than run standalone.
# Done via Python, not sed: sed's in-place-edit flag differs between BSD
# sed (macOS) and GNU sed (Linux) -- "sed -i ''" only works on BSD and
# fails outright on Linux (proven by running it), which would break this
# script anywhere but a Mac. Python's text handling is identical everywhere.
uv run python -c "
import pathlib
path = pathlib.Path('app/pb/analysis_pb2_grpc.py')
content = path.read_text()
old = 'import analysis_pb2 as analysis__pb2'
new = 'from app.pb import analysis_pb2 as analysis__pb2'
if old not in content:
    raise SystemExit(
        f'Expected import line not found in {path} -- '
        'protoc output format may have changed, check manually.'
    )
path.write_text(content.replace(old, new))
"

echo "Generated protobuf files:"
echo "  app/pb/analysis_pb2.py"
echo "  app/pb/analysis_pb2.pyi"
echo "  app/pb/analysis_pb2_grpc.py"
echo "  app/pb/analysis_pb2_grpc.pyi"
echo
echo "Verifying generated imports..."
grep -E '^import .*pb2|^from .*pb2' app/pb/analysis_pb2_grpc.py
echo
echo "Testing Python imports..."
uv run python -c "from app.pb import analysis_pb2, analysis_pb2_grpc; print('Proto imports: OK')"
echo
echo "Proto generation complete."
