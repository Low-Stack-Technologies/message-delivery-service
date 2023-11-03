# POST /v2/sms

## Headers

```json
{
  "Authorization": "Bearer {{JWT TOKEN}}",
  "Content-Type": "application/json"
}
```

## Body (Without templating)

```json
{
  "to": "{{PHONE NUMBER}}",
  "from": {
    "name": "{{NAME}}"
  },

  "body": "{{BODY}}"
}
```

## Body (With templating)

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

## Response

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

## Response (Error)

```json
{
  "success": false,
  "message": "Invalid phone number"
}
```
