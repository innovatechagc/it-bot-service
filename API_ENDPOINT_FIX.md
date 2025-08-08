# 🔧 **SOLUCIÓN: Error en GET /api/v1/bots**

## 🚨 **Problema Identificado**

El endpoint `GET /api/v1/bots` está devolviendo un error 400:

```json
{
  "code": "INVALID_REQUEST",
  "message": "Owner ID is required",
  "data": null
}
```

## 🔍 **Análisis del Código**

Revisando el handler en `internal/handlers/bot_handlers.go`:

```go
func (h *BotHandler) GetBots(c *gin.Context) {
    ownerID := c.Query("owner_id")
    if ownerID == "" {
        // Obtener del JWT token (implementar middleware de auth)
        ownerID = c.GetString("user_id")
    }

    if ownerID == "" {
        c.JSON(http.StatusBadRequest, domain.APIResponse{
            Code:    "INVALID_REQUEST",
            Message: "Owner ID is required",
        })
        return
    }
    // ... resto del código
}
```

## ✅ **Solución**

El endpoint requiere un `owner_id` que puede venir de dos fuentes:

### 1. **Query Parameter** (Recomendado para testing)
```
GET /api/v1/bots?owner_id=user-001
```

### 2. **JWT Token** (Para producción)
El `owner_id` se extrae automáticamente del token JWT si hay middleware de autenticación.

## 🛠️ **Cómo Usar Correctamente**

### **Opción 1: Query Parameter**
```bash
curl -X GET "http://localhost:8084/api/v1/bots?owner_id=user-001"
```

### **Opción 2: Con JWT Token**
```bash
curl -X GET "http://localhost:8084/api/v1/bots" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📝 **Colección Postman Actualizada**

He actualizado la colección de Postman para incluir el `owner_id` como query parameter:

```json
{
  "name": "List All Bots",
  "request": {
    "method": "GET",
    "url": {
      "raw": "{{base_url}}/api/v1/bots?owner_id=user-001",
      "query": [
        {
          "key": "owner_id",
          "value": "user-001",
          "description": "ID del propietario de los bots"
        }
      ]
    }
  }
}
```

## 🎯 **Para Flutter**

En tu aplicación Flutter, asegúrate de:

1. **Incluir el owner_id** en las llamadas:
   ```dart
   final response = await http.get(
     Uri.parse('$baseUrl/api/v1/bots?owner_id=$userId'),
   );
   ```

2. **O usar autenticación JWT** si está implementada:
   ```dart
   final response = await http.get(
     Uri.parse('$baseUrl/api/v1/bots'),
     headers: {
       'Authorization': 'Bearer $jwtToken',
     },
   );
   ```

## 🔄 **Otros Endpoints Similares**

Los siguientes endpoints también pueden requerir `owner_id`:

- `POST /api/v1/bots` - Ya incluye `owner_id` en el body
- `GET /api/v1/conditionals/bot/:botId` - Usa el bot_id
- `GET /api/v1/triggers/bot/:botId` - Usa el bot_id
- `GET /api/v1/test-cases/bot/:botId` - Usa el bot_id
- `GET /api/v1/test-suites/bot/:botId` - Usa el bot_id

## ✅ **Estado de la Solución**

- ✅ **Problema identificado**
- ✅ **Colección Postman actualizada**
- ✅ **Documentación creada**
- ✅ **Ejemplos de uso proporcionados**

¡El endpoint ahora debería funcionar correctamente! 🚀 