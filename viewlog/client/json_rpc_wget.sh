wget --header="Content-Type: application/json" \
	--header="Accept: application/json" \
	--post-data='{"jsonrpc":"2.0","method":"Sessions.Length","params":null,"id":1}' \
	--no-check-certificate \
	http://localhost:8888/debug/rpc
