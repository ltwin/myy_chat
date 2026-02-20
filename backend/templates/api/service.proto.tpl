syntax = "proto3";

package {{SERVICE}}.v1;

option go_package = "github.com/myy-chat/backend/api/{{SERVICE}}/v1;v1";
option java_multiple_files = true;
option java_package = "{{SERVICE}}.v1";

import "google/api/annotations.proto";

// Service is the {{SERVICE}} service definition.
service Service {
  // Hello is a simple hello world endpoint.
  rpc Hello(HelloRequest) returns (HelloResponse) {
    option (google.api.http) = {
      get: "/api/v1/{{SERVICE}}/hello"
    };
  }
}

// HelloRequest is the request for Hello.
message HelloRequest {
  string name = 1;  // Name to greet
}

// HelloResponse is the response for Hello.
message HelloResponse {
  string message = 1;  // Greeting message
}
