# Message Delivery Service

This is a service to easily provide email and SMS sending capabilities to other services through a simple REST API. It is built using Node.js and Express along with Zod for validation.

# Docker Compose

The easiest way to run this service is through Docker Compose. The following is an example `docker-compose.yml` file:

```yaml
services:
  message-delivery-service:
    image: dotzerotechnologies/message-delivery-service:latest
    ports:
      - 3000:3000
    environment:
      - DATA_PATH=/data
      - CONFIGURATION_FILE_PATH=/data/config.json
    volumes:
      - ./data:/data
    restart: unless-stopped
```

# Documentation

## POST /v2/authenticate

### Headers

```json
{
  "Content-Type": "application/json"
}
```

### Body

```json
{
  "service": "{{SERVICE NAME}}",
  "password": "{{PASSWORD}}",
  "totp": "{{TOTP CODE}}"
}
```

### Response

```json
{
  "success": true,
  "message": "Authenticated successfully",
  "data": {
    "jwt": "{{JWT TOKEN}}"
  }
}
```

### Response (Error)

```json
{
  "success": false,
  "message": "Invalid credentials"
}
```

## POST /v2/email

### Headers

```json
{
  "Authorization": "Bearer {{JWT TOKEN}}",
  "Content-Type": "application/json"
}
```

### Body (Without templating)

```json
{
  "to": "{{EMAIL ADDRESS}}",
  "from": {
    "name": "{{NAME}}",
    "email": "{{EMAIL ADDRESS}}"
  },
  "subject": "{{SUBJECT}}",

  "useTemplate": false,
  "body": "{{BODY}}",
  "isHTML": true // or false
}
```

### Body (With templating)

```json
{
  "to": "{{EMAIL ADDRESS}}",
  "from": {
    "name": "{{NAME}}",
    "email": "{{EMAIL ADDRESS}}"
  },
  "subject": "{{SUBJECT}}",

  "useTemplate": true,
  "template": {
    "name": "{{TEMPLATE NAME}}",
    "data": {
      "{{KEY}}": "{{VALUE}}"
      // ...
    }
  }
}
```

### Response

```json
{
  "success": true,
  "message": "Email sent successfully"
}
```

### Response (Error)

```json
{
  "success": false,
  "message": "Error message"
}
```

## POST /v2/sms

### Headers

```json
{
  "Authorization": "Bearer {{JWT TOKEN}}",
  "Content-Type": "application/json"
}
```

### Body (Without templating)

```json
{
  "to": {
    "phone": "{{PHONE NUMBER}}",
    "country": "{{COUNTRY CODE}}" // Two-letter country code
  },
  "from": {
    "name": "{{NAME}}"
  },

  "useTemplate": false,
  "body": "{{BODY}}"
}
```

### Body (With templating)

```json
{
  "to": "{{PHONE NUMBER}}",
  "from": {
    "name": "{{NAME}}"
  },

  "useTemplate": true,
  "template": {
    "name": "{{TEMPLATE NAME}}",
    "data": {
      "{{KEY}}": "{{VALUE}}"
      // ...
    }
  }
}
```

### Response

```json
{
  "success": true,
  "message": "SMS sent successfully",
  "cost": {
    "currency": "SEK",
    "amount": 0.35 // Cost in 10000ths of the currency
  }
}
```

### Response (Error)

```json
{
  "success": false,
  "message": "Invalid phone number"
}
```
