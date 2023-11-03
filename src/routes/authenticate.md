# POST /v2/authenticate

## Headers

```json
{
  "Content-Type": "application/json"
}
```

## Body

```json
{
  "service": "{{SERVICE NAME}}",
  "password": "{{PASSWORD}}",
  "totp": "{{TOTP CODE}}"
}
```

## Response

```json
{
  "success": true,
  "message": "Authenticated successfully",
  "data": {
    "jwt": "{{JWT TOKEN}}"
  }
}
```

## Response (Error)

```json
{
  "success": false,
  "message": "Invalid credentials"
}
```
