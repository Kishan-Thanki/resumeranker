#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Generating Python gRPC code..."

uv run python -m grpc_tools.protoc \
  -I ../proto \
  --python_out=src/analysis_dashboard/pb \
  --grpc_python_out=src/analysis_dashboard/pb \
  --mypy_out=src/analysis_dashboard/pb \
  --mypy_grpc_out=src/analysis_dashboard/pb \
  ../proto/analysis.proto

# grpc_tools.protoc generates:
#
#   import analysis_pb2 as analysis__pb2
#
# inside analysis_pb2_grpc.py. Since the generated files live inside
# the analysis_dashboard.pb package, use a package-qualified import instead.
uv run python -c "
from pathlib import Path

path = Path('src/analysis_dashboard/pb/analysis_pb2_grpc.py')
content = path.read_text()

old = 'import analysis_pb2 as analysis__pb2'
new = 'from analysis_dashboard.pb import analysis_pb2 as analysis__pb2'

if old not in content:
    raise SystemExit(
        f'Expected import line not found in {path}. '
        'The protoc output format may have changed.'
    )

path.write_text(content.replace(old, new))
"

echo "Generated protobuf files:"
echo "  src/analysis_dashboard/pb/analysis_pb2.py"
echo "  src/analysis_dashboard/pb/analysis_pb2.pyi"
echo "  src/analysis_dashboard/pb/analysis_pb2_grpc.py"
echo "  src/analysis_dashboard/pb/analysis_pb2_grpc.pyi"

echo
echo "Testing protobuf imports..."

uv run python -c "
from analysis_dashboard.pb import analysis_pb2
from analysis_dashboard.pb import analysis_pb2_grpc

print('Proto imports: OK')
"

echo
echo "Proto generation complete."
