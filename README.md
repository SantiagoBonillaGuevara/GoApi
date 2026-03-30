# 🔐 JWT Auth API — Go + AWS Lambda

REST API de autenticación con JWT construida en Go, optimizada para tiempos de arranque en frío mínimos en AWS Lambda.

---

## 📋 Descripción

Este proyecto implementa un sistema de autenticación basado en **JWT (JSON Web Tokens)** con control de acceso por roles. Está diseñado para correr como función **AWS Lambda** aprovechando que el runtime de Go tiene uno de los tiempos de arranque en frío más bajos disponibles (~5–15 ms), lo que lo hace ideal para workloads serverless con latencia crítica.

---

## ⚡ ¿Por qué Go en Lambda para cold starts rápidos?

| Runtime       | Cold Start típico |
|---------------|-------------------|
| **Go**        | ~5–15 ms          |
| Node.js       | ~100–200 ms       |
| Python        | ~100–300 ms       |
| Java          | ~500–1000 ms      |
| .NET          | ~200–500 ms       |

Go compila a un **binario nativo** sin JVM ni intérprete, lo que elimina la sobrecarga de inicialización. Combinado con Lambda, el resultado es una API extremadamente rápida desde la primera invocación.

---

## 🏗️ Arquitectura

```
Cliente HTTP
    │
    ▼
AWS API Gateway
    │
    ▼
AWS Lambda (Go binary)
    │
    ├── POST /users/login   → Genera JWT
    ├── GET  /users/a       → Requiere rol A
    ├── GET  /users/b       → Requiere rol B
    └── GET  /users/c       → Requiere rol C
```

---

## 📁 Estructura del Proyecto

```
.
├── main.go          # Punto de entrada, handlers, middleware y rutas
├── go.mod           # Módulo y dependencias
├── go.sum           # Hashes de dependencias
└── README.md
```

---

## 🔧 Dependencias

| Paquete | Versión | Uso |
|--------|---------|-----|
| `github.com/gorilla/mux` | v1.8.1 | Router HTTP |
| `github.com/dgrijalva/jwt-go` | v3.2.0 | Generación y validación de JWT |

> **Nota:** Para producción se recomienda migrar `dgrijalva/jwt-go` a [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt), ya que el primero no tiene mantenimiento activo.

---

## 🚀 Instalación y Ejecución Local

### Prerrequisitos

- Go 1.24.1+
- AWS CLI (para despliegue)

### Clonar y ejecutar

```bash
git clone <repo-url>
cd api-go

# Descargar dependencias
go mod tidy

# Ejecutar localmente
go run main.go
# → Servidor corriendo en :8080
```

---

## 📡 Endpoints

### `POST /users/login`

Autentica al usuario y retorna un JWT.

**Request:**
```json
{
  "name": "userA",
  "password": "userA"
}
```

**Response `200 OK`:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Errores:**
- `400 Bad Request` — cuerpo inválido
- `401 Unauthorized` — credenciales incorrectas

---

### `GET /users/a` · `GET /users/b` · `GET /users/c`

Endpoints protegidos por rol. Requieren el header `Authorization: Bearer <token>`.

**Roles disponibles:**

| Endpoint    | Rol requerido |
|-------------|---------------|
| `/users/a`  | `A`           |
| `/users/b`  | `B`           |
| `/users/c`  | `C`           |

**Headers:**
```
Authorization: Bearer <JWT>
```

**Response `200 OK`:**
```
Solo el rol A puede acceder aquí
```

**Errores:**
- `401 Unauthorized` — token ausente o inválido
- `403 Forbidden` — rol incorrecto

---

## 🔑 Estructura del JWT

```json
{
  "sub":   "userA",
  "roles": "A",
  "iss":   "api-go",
  "iat":   1710000000,
  "exp":   1710003600
}
```

El token expira **1 hora** después de su emisión. Firmado con **HMAC-SHA256**.

---

## ☁️ Despliegue en AWS Lambda

### 1. Compilar para Lambda (Linux AMD64)

```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip function.zip bootstrap
```

> El binario se llama `bootstrap` para usar el runtime `provided.al2` de AWS.

### 2. Crear la función Lambda

```bash
aws lambda create-function \
  --function-name jwt-api-go \
  --runtime provided.al2 \
  --role arn:aws:iam::<ACCOUNT_ID>:role/<LAMBDA_ROLE> \
  --handler bootstrap \
  --zip-file fileb://function.zip
```

### 3. Configurar API Gateway

Conectar AWS API Gateway (HTTP API o REST API) a la función Lambda para exponer los endpoints públicamente.

### 4. Variables de entorno recomendadas

| Variable | Descripción |
|----------|-------------|
| `SECRET_KEY` | Clave secreta para firmar JWT (no hardcodear en código) |

> ⚠️ En producción, mueve `secretKey` a una variable de entorno o AWS Secrets Manager.

---

## 👥 Usuarios de Prueba

| Usuario | Contraseña | Rol |
|---------|-----------|-----|
| userA   | userA     | A   |
| userB   | userB     | B   |
| userC   | userC     | C   |

> ⚠️ Estos usuarios están quemados en el código. En producción, reemplaza por una base de datos o un Identity Provider (Cognito, Auth0, etc.).

---

## 🔒 Consideraciones de Seguridad

- [ ] Rotar `SECRET_KEY` y externalizarla (env var / Secrets Manager)
- [ ] Migrar a `golang-jwt/jwt` (fork mantenido)
- [ ] Reemplazar usuarios hardcodeados por una fuente de datos real
- [ ] Habilitar HTTPS (API Gateway lo maneja automáticamente)
- [ ] Considerar refresh tokens para sesiones de larga duración
- [ ] Agregar rate limiting en API Gateway

---

## 📄 Licencia

MIT
