#!/bin/sh
# Each Go step recompiles the embeds produced by the previous step.
set -eu
cd "$(dirname "$0")"
module_root=$(cd ../../.. && pwd)
uv sync --locked --project "$module_root"
python="$module_root/.venv/bin/python"

"$python" -m unittest test_mapping_generators.py
"$python" export_defaults.py
"$python" gen_fan_in_mapping.py
"$python" gen_fan_out_mapping.py
"$python" gen_java_compat.py

for step in \
    TestStep0ReadBlocksFromPalette \
    TestReadDefaultsReadBlocksConsistency \
    TestStep2_1_GenBedrockFanin \
    TestStep2_2_GenJavaDefaultPropFanin \
    TestStep2_3_GenJavaFanin \
    TestStep3GenJavaFanout \
    TestStep4BuildBlockDataByName \
    TestSchematic
do
    go test -tags resourcegen -v -count=1 -run "^$step\$" .
done
go test -tags resourcegen -v -count=1 \
    -run '^(TestReadDefaultsAllJSON|TestStep4StorageDumpLoad)$' .
go test ../../../presets/...
