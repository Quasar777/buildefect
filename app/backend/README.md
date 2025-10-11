# BuildDefect API Documentation

**Base URL:** `/api`

**Auth:** JWT Bearer Token (`Authorization: Bearer <token>`)

**Format:** JSON

**Error format:**

```json
{
  "error": "описание ошибки"
}
```

**Common HTTP status codes:**

* `200 OK` — успешный запрос
* `201 Created` — успешно создан ресурс
* `400 Bad Request` — некорректные данные запроса
* `401 Unauthorized` — неавторизован
* `403 Forbidden` — недостаточно прав
* `404 Not Found` — ресурс не найден
* `500 Internal Server Error` — внутренняя ошибка сервера

---

## 1. Auth (Авторизация)

### 1.1 Регистрация

**POST** `/auth/register`
**Body:**

```json
{
  "login": "user1",
  "password": "pass123",
  "name": "Иван",
  "lastname": "Иванов"
}
```

**Response 201:**

```json
{
  "id": 1,
  "login": "user1",
  "name": "Иван",
  "lastname": "Иванов",
  "role": "engineer"
}
```

**Errors:** `400`, `409`, `500`

---

### 1.2 Логин

**POST** `/auth/login`
**Body:**

```json
{
  "login": "user1",
  "password": "pass123"
}
```

**Response 200:**

```json
{
  "access_token": "jwt.token.here",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Errors:** `400`, `401`, `500`

---

## 2. Buildings (Здания)

### 2.1 Создать здание

**POST** `/buildings`
**Body:**

```json
{
  "name": "ЖК Солнечный",
  "address": "ул. Ленина, 1",
  "stage": "строительство"
}
```

**Response 201:**

```json
{
  "id": 1,
  "name": "ЖК Солнечный",
  "address": "ул. Ленина, 1",
  "stage": "строительство"
}
```

### 2.2 Получить все здания

**GET** `/buildings`
**Response 200:** массив объектов `BuildingResponse`

### 2.3 Получить здание по ID

**GET** `/buildings/{id}`
**Response 200:** объект `BuildingResponse`
**Errors:** `400`, `404`, `500`

### 2.4 Обновить здание

**PATCH** `/buildings/{id}`
**Body:** любое сочетание полей `name`, `address`, `stage`
**Response 200:** обновлённый объект `BuildingResponse`
**Errors:** `400`, `404`, `500`

### 2.5 Удалить здание

**DELETE** `/buildings/{id}`
**Response 200:** `"Successfully deleted building with id {id}"`
**Errors:** `400`, `404`, `500`

---

## 3. Defects (Дефекты)

### 3.1 Создать дефект

**POST** `/defects`
**Body:**

```json
{
  "building_id": 1,
  "title": "Протечка крыши",
  "description": "После дождя вода попадает в квартиру",
  "priority": "high",
  "responsible_person_id": 2,
  "deadline": "2025-10-20 12:00:00",
  "status": "new"
}
```

**Response 201:** объект `DefectResponse`
**Errors:** `400`, `401`, `500`

### 3.2 Получить список дефектов

**GET** `/defects`
**Query params:** `status`, `building_id`, `responsible_id`, `limit`, `offset`
**Response 200:** массив объектов `DefectResponse`

### 3.3 Получить дефект по ID

**GET** `/defects/{id}`
**Response 200:** объект `DefectResponse`
**Errors:** `400`, `404`, `500`

### 3.4 Удалить дефект

**DELETE** `/defects/{id}`
**Response 200:** `"Successfully deleted defect with id {id}"`
**Errors:** `400`, `404`, `500`

---

## 4. Comments (Комментарии)

### 4.1 Создать комментарий

**POST** `/comments`
**Body:**

```json
{
  "defect_id": 1,
  "text": "Комментарий по дефекту"
}
```

**Response 201:** объект `CommentResponse`
**Errors:** `400`, `401`, `500`

### 4.2 Получить комментарии по дефекту

**GET** `/comments?defect_id={id}`
**Response 200:** массив объектов `CommentResponse`
**Errors:** `400`, `500`

### 4.3 Получить комментарий по ID

**GET** `/comments/{id}`
**Response 200:** объект `CommentResponse`
**Errors:** `400`, `404`, `500`

### 4.4 Удалить комментарий

**DELETE** `/comments/{id}` — только для пользователей с ролью `observer`
**Response 200:** `"Successfully deleted comment with id {id}"`
**Errors:** `400`, `403`, `404`, `500`

---

## 5. Defect Attachments (Вложения к дефектам)

### 5.1 Загрузить файл

**POST** `/defect_attachments/{defect_id}`
**Body:** multipart/form-data, поле `file`
**Response 201:** объект `DefectAttachment`
**Errors:** `400`, `404`, `500`

### 5.2 Получить список вложений дефекта

**GET** `/defect_attachments/{defect_id}`
**Response 200:** массив объектов `DefectAttachment`
**Errors:** `400`, `404`, `500`

### 5.3 Получить вложение по ID

**GET** `/defect_attachments/file/{id}`
**Response 200:** объект `DefectAttachment`
**Errors:** `400`, `404`, `500`

### 5.4 Удалить вложение

**DELETE** `/defect_attachments/file/{id}`
**Response 200:** `{"message": "attachment deleted successfully"}`
**Errors:** `400`, `404`, `500`


## 6. Общие рекомендации для фронтенда

* Всегда проверять коды ошибок и выводить пользователю понятные сообщения
* Использовать `Bearer Token` для всех авторизованных операций
* Формат даты и времени: `"YYYY-MM-DD HH:mm:ss"`
* Пагинация для дефектов: `limit` и `offset`
* Все объекты имеют уникальный `id` для ссылок и обновлений

