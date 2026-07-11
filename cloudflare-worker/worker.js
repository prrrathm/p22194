export default {
  async fetch(request) {
    const CLOUD_RUN_URL = "https://api-gateway-3b3mud4qva-em.a.run.app";
    const url = new URL(request.url);
    const target = CLOUD_RUN_URL + url.pathname + url.search;

    // Handle CORS preflight
    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS",
          "Access-Control-Allow-Headers": "Authorization, Content-Type, X-Request-ID, X-Session-Token",
          "Access-Control-Max-Age": "86400",
        },
      });
    }

    // Proxy request to Cloud Run
    const proxyRequest = new Request(target, {
      method: request.method,
      headers: request.headers,
      body: request.body,
    });

    const response = await fetch(proxyRequest);

    // Add CORS headers to response
    const newResponse = new Response(response.body, response);
    newResponse.headers.set("Access-Control-Allow-Origin", "*");
    newResponse.headers.set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS");
    newResponse.headers.set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Session-Token");

    return newResponse;
  },
};
