[![godoc](https://godoc.org/github.com/4freewifi/smshinet?status.svg)](https://godoc.org/github.com/4freewifi/smshinet)


# HiNet SMS

Go implementation of HiNet SMS service client, including both library
and a standalone JSON-RPC service proxy.

For complete document, check
https://godoc.org/github.com/4freewifi/smshinet .

# Example

* `unit_test.go` should give a pretty good idea about library usage.
* Check `contrib/jsonrpc.py` regarding how to use the standalone
  JSON-RPC server.

# References

1. [軟體規格](https://sms.hinet.net/new/sent_software.htm)
2. [程式開發](https://sms.hinet.net/new/sent_program.htm)

# License

Copyright 2015 John Lee <john@4free.com.tw>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
