# #!/bin/bash

# HOST="127.0.0.1"
# PORT="3737"

# PAYLOAD=$(cat <<EOF
# {
#   "jsonrpc": "2.0",
#   "id": 1,
#   "method": "initialize",
#   "params": {
#     "processId": null,
#     "rootUri": "file:///tmp",
#     "capabilities": {},
#     "trace": "off",
#     "workspaceFolders": null
#   }
# }
# EOF
# )

# # Hitung panjang payload
# LENGTH=$(echo -n "$PAYLOAD" | wc -c)

# # Gabungkan header dan body dengan \r\n sesuai LSP spec
# {
#   echo -en "Content-Length: $LENGTH\r\n\r\n$PAYLOAD"
#   sleep 1  # penting: biar nc gak langsung close, kasih waktu untuk read response
# } | nc $HOST $PORT

#!/bin/bash

HOST="127.0.0.1"
PORT="3737"

send_lsp() {
  local payload="$1"
  local length=$(echo -n "$payload" | wc -c)
  echo -ne "Content-Length: $length\r\n\r\n$payload" | nc $HOST $PORT
  sleep 1  # penting: biar nc gak langsung close, kasih waktu untuk read response
}

# 1. Initialize
send_lsp '
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "processId": null,
    "rootUri": "file:///Users/ibanrama-master/Documents/Developments/OpenSourceProject/ExecLSP/backend",
    "capabilities": {},
    "trace": "off",
    "workspaceFolders": null
  }
}'

sleep 1

# 2. Send initialized notification
send_lsp '{
  "jsonrpc": "2.0",
  "method": "initialized",
  "params": {}
}'

sleep 1

# 3. Send didOpen for /tmp/main.go
send_lsp '{
  "jsonrpc": "2.0",
  "method": "textDocument/didOpen",
  "params": {
    "textDocument": {
      "uri": "file:///Users/ibanrama-master/Documents/Developments/OpenSourceProject/ExecLSP/backend/main.go",
      "languageId": "go",
      "version": 1,
      "text": "package main\n\nfunc main() {\n  println(\"hello\")\n}\n"
    }
  }
}'

sleep 1

# 4. Send hover request di posisi "println"
send_lsp '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "textDocument/hover",
  "params": {
    "textDocument": {
      "uri": "file:///tmp/main.go"
    },
    "position": {
      "line": 3,
      "character": 3
    }
  }
}'
