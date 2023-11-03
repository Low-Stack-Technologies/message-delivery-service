# POST /v2/email

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

## Body (With templating)

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

## Response

```json
{
  "success": true,
  "message": "Email sent successfully"
}
```

## Response (Error)

```json
{
  "success": false,
  "message": "Error message"
}
```
