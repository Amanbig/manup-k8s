# Application Integration Guide

When using **ManUp Injector**, secrets are mounted inside your container in an in-memory `.env` file (by default at `/app/secrets/.env`). Here is how applications in different programming languages load them.

---

## 🟢 Node.js / TypeScript

### Option A: Standard `dotenv`
```typescript
import dotenv from 'dotenv';
import path from 'path';

// Specify the path to the injected RAM secrets
const secretPath = process.env.SECRETS_FILE || '/app/secrets/.env';
dotenv.config({ path: secretPath });

console.log('Database password loaded:', process.env.DATABASE_PASSWORD);
```

### Option B: Preload via CLI Flag
No code changes required! Adjust your container start command:
```json
{
  "scripts": {
    "start": "node -r dotenv/config dist/index.js dotenv_config_path=/app/secrets/.env"
  }
}
```

---

## 🐍 Python

### Using `python-dotenv`
```python
import os
from dotenv import load_dotenv

# Load secrets from the injected in-memory volume
load_dotenv('/app/secrets/.env')

db_password = os.getenv('DATABASE_PASSWORD')
api_token = os.getenv('STRIPE_API_KEY')
```

---

## 🔵 Go (Golang)

### Using `godotenv`
```go
package main

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

func main() {
    // Load injected secrets
    if err := godotenv.Load("/app/secrets/.env"); err != nil {
        log.Printf("Warning: /app/secrets/.env not found, using OS env vars")
    }

    dbPass := os.Getenv("DATABASE_PASSWORD")
    _ = dbPass
}
```

---

## ☕ Java / Spring Boot

Spring Boot natively supports loading configuration from environment files or directories:

```yaml
# In your application.yml or command arguments
# Pass --spring.config.additional-location=file:/app/secrets/
```

Or pass via entrypoint wrapper:
```sh
sh -c '. /app/secrets/.env && exec java -jar app.jar'
```

---

## 🐚 Generic Shell / Any Binary

Wrap your container's entrypoint to source the `.env` file before executing your application binary. This exports every variable into `process.env`:

```yaml
spec:
  containers:
    - name: app
      image: my-app:latest
      command: ["/bin/sh", "-c"]
      args:
        - |
          if [ -f /app/secrets/.env ]; then
            set -a
            . /app/secrets/.env
            set +a
          fi
          exec ./my-app-binary
```
