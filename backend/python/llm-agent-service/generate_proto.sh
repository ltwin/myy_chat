#!/bin/bash
# Generate Python gRPC code from proto files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROTO_DIR="${SCRIPT_DIR}/proto"
OUTPUT_DIR="${SCRIPT_DIR}/app/api/generated"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Generate Python gRPC code
python -m grpc_tools.protoc \
  -I"${PROTO_DIR}" \
  --python_out="${OUTPUT_DIR}" \
  --grpc_python_out="${OUTPUT_DIR}" \
  --pyi_out="${OUTPUT_DIR}" \
  "${PROTO_DIR}/llm_agent.proto"

# Create __init__.py
touch "${OUTPUT_DIR}/__init__.py"

echo "Proto files generated successfully in ${OUTPUT_DIR}"
