# Todo REST API

A simple REST API for managing todo items built with Go and MongoDB.

## Features

- Create, read, update, and delete todo items
- MongoDB integration
- RESTful API design
- CORS support
- JSON responses

## Prerequisites

- Go 1.21 or higher
- MongoDB running on localhost:27017

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Make sure MongoDB is running on localhost:27017
4. Run the application:
   ```bash
   go run main.go
   ```

The server will start on port 8080.

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### Get All Todos
```
GET /todos
```
Returns an array of all todo items.

**Response:**
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "title": "Sample Todo",
    "description": "This is a sample todo item",
    "completed": false,
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z"
  }
]
```

#### Get Single Todo
```
GET /todos/{id}
```
Returns a specific todo item by ID.

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "title": "Sample Todo",
  "description": "This is a sample todo item",
  "completed": false,
  "created_at": "2023-12-01T10:00:00Z",
  "updated_at": "2023-12-01T10:00:00Z"
}
```

#### Create Todo
```
POST /todos
```
Creates a new todo item.

**Request Body:**
```json
{
  "title": "New Todo",
  "description": "Description of the new todo",
  "completed": false
}
```

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "title": "New Todo",
  "description": "Description of the new todo",
  "completed": false,
  "created_at": "2023-12-01T10:00:00Z",
  "updated_at": "2023-12-01T10:00:00Z"
}
```

#### Update Todo
```
PUT /todos/{id}
```
Updates an existing todo item.

**Request Body:**
```json
{
  "title": "Updated Todo",
  "description": "Updated description",
  "completed": true
}
```

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "title": "Updated Todo",
  "description": "Updated description",
  "completed": true,
  "created_at": "2023-12-01T10:00:00Z",
  "updated_at": "2023-12-01T10:30:00Z"
}
```

#### Delete Todo
```
DELETE /todos/{id}
```
Deletes a todo item.

**Response:** 204 No Content

## Error Responses

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request` - Invalid request data
- `404 Not Found` - Todo item not found
- `500 Internal Server Error` - Server error

## Example Usage with curl

### Create a todo
```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","description":"Study Go programming language","completed":false}'
```

### Get all todos
```bash
curl http://localhost:8080/api/v1/todos
```

### Update a todo
```bash
curl -X PUT http://localhost:8080/api/v1/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","description":"Study Go programming language","completed":true}'
```

### Delete a todo
```bash
curl -X DELETE http://localhost:8080/api/v1/todos/{id}
```

## Database

The application uses MongoDB with the following configuration:
- Database: `todoapp`
- Collection: `todos`
- Connection: `mongodb://localhost:27017`

## Todo Schema

```go
type Todo struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Title       string             `json:"title" bson:"title"`
    Description string             `json:"description" bson:"description"`
    Completed   bool               `json:"completed" bson:"completed"`
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
```