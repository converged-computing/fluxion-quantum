#!/usr/bin/env bash
# Diagnose the 400 from acquire(): exchange apikey -> bearer token, then try to
# create a session exactly like QRMI does, printing IBM's response body.
set -u
: "${IBM_CLOUD_TOKEN:?set IBM_CLOUD_TOKEN (your IAM API key)}"
: "${IBM_CLOUD_CRN:?set IBM_CLOUD_CRN (your Qiskit Runtime instance CRN)}"
BACKEND="${QRMI_TEST_BACKEND:-ibm_fez}"
QRS="${QRMI_QRS_ENDPOINT:-https://quantum.cloud.ibm.com/api/v1}"
IAM="${QRMI_QRS_IAM_ENDPOINT:-https://iam.cloud.ibm.com}"

echo "== exchanging API key for a bearer token at $IAM =="
TOKEN=$(curl -s -X POST "$IAM/identity/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=${IBM_CLOUD_TOKEN}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin).get("access_token",""))')
if [ -z "$TOKEN" ]; then echo "IAM token exchange failed (check the API key)"; exit 1; fi
echo "got token (${#TOKEN} chars)"

for MODE in dedicated batch; do
  echo
  echo "== POST $QRS/sessions  (mode=$MODE, backend=$BACKEND, max_ttl=28800) =="
  curl -s -i -X POST "$QRS/sessions" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Service-CRN: ${IBM_CLOUD_CRN}" \
    -H "Content-Type: application/json" \
    -d "{\"backend\":\"${BACKEND}\",\"mode\":\"${MODE}\",\"max_ttl\":28800}" \
    | sed -n '1,40p'
done