// Propagate Set-Cookie headers from subgraphs to client and set HTTP status codes
fn supergraph_service(service) {
  let process_response = |response| {
    // Propagate Set-Cookie headers
    if response.context["set_cookie_headers"]?.len > 0 {
      response.headers["set-cookie"] = response.context["set_cookie_headers"];
    }

    // Set HTTP status code based on GraphQL errors
    if response.body.errors != () && response.body.errors.len() > 0 {
      let error = response.body.errors[0];

      // Check for error code in extensions
      if error.extensions != () {
        let code = error.extensions.code;

        if code == "UNAUTHENTICATED" {
          response.status_code = 401;
        } else if code == "FORBIDDEN" {
          response.status_code = 403;
        } else if code == "BAD_REQUEST" || code == "GRAPHQL_VALIDATION_FAILED" {
          response.status_code = 400;
        } else if code == "INTERNAL_SERVER_ERROR" {
          response.status_code = 500;
        }
      } else {
        // Check error message for authentication/authorization
        let message = error.message;
        if message != () {
          if message.contains("not authenticated") || message.contains("unauthorized") {
            response.status_code = 401;
          } else if message.contains("forbidden") || message.contains("access denied") {
            response.status_code = 403;
          } else if message.contains("bad request") || message.contains("invalid") {
            response.status_code = 400;
          } else {
            response.status_code = 500;
          }
        }
      }
    }
  };

  service.map_response(process_response);
}

fn subgraph_service(service, subgraph) {
  // Store Set-Cookie headers from subgraph responses
  let store_cookies_from_subgraphs = |response| {
    if "set-cookie" in response.headers {
      if response.context["set_cookie_headers"] == () {
        response.context.set_cookie_headers = []
      }

      response.context.set_cookie_headers += response.headers.values("set-cookie");
    }
  };

  service.map_response(store_cookies_from_subgraphs);
}
