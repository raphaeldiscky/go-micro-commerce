#!/usr/bin/env node

// Simple JWT token generator for testing
const crypto = require("crypto");

function base64urlEscape(str) {
  return str.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

function base64urlEncode(str) {
  return base64urlEscape(Buffer.from(str).toString("base64"));
}

function sign(message, secret) {
  return base64urlEscape(
    crypto.createHmac("sha256", secret).update(message).digest("base64")
  );
}

// JWT Header
const header = {
  alg: "HS256",
  typ: "JWT",
};

// JWT Payload
const payload = {
  sub: "test-user-id",
  user_id: "test-user-id",
  email: "test@example.com",
  roles: ["user"],
  is_active: true,
  iss: "api-gateway",
  exp: Math.floor(Date.now() / 1000) + 24 * 60 * 60, // 24 hours from now
  iat: Math.floor(Date.now() / 1000),
};

const secret = "your-256-bit-secret-key-here-change-in-production";

const encodedHeader = base64urlEncode(JSON.stringify(header));
const encodedPayload = base64urlEncode(JSON.stringify(payload));
const signature = sign(`${encodedHeader}.${encodedPayload}`, secret);

const token = `${encodedHeader}.${encodedPayload}.${signature}`;

console.log("JWT Token:");
console.log(token);
console.log("\nTo use in curl:");
console.log(
  `curl -H "Authorization: Bearer ${token}" http://localhost:8080/api/v1/products`
);
