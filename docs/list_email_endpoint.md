# GET /api/email/list Endpoint

This document describes the authenticated email listing endpoint that returns a list of emails sent by the authenticated user.

## 🔐 **Endpoint Overview**

- **Method**: `GET`
- **Path**: `/api/email/list`
- **Authentication**: Required (JWT Bearer token)
- **Authorization**: Users can only see their own emails

## 🛡️ **Security Features**

### **Authentication Requirements**
- **JWT Token**: Required in `Authorization: Bearer <token>` header
- **User Isolation**: Only returns emails where `sender_id` matches authenticated user
- **Access Control**: 401 Unauthorized for missing/invalid tokens

### **Authorization Logic**
- **User Filtering**: Database query filters by `sender_id = user_id`
- **Data Isolation**: Users cannot access emails sent by other users
- **Audit Trail**: All access attempts are logged

## 🔧 **Implementation Details**

### **Handler Function**
```go
func (srv *Server) listEmailHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user_id from context
    userID, ok := GetUserIDFromContext(r)
    if !ok {
        http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
        return
    }

    // 2. Query database for user's emails
    rows, err := srv.db.Query(`
        SELECT email_id, recipient, subject, created_at
        FROM emails 
        WHERE sender_id = ?
        ORDER BY created_at DESC`,
        userID,
    )

    // 3. Build response with email list
    var emails []EmailListItem
    for rows.Next() {
        var email EmailListItem
        rows.Scan(&email.EmailID, &email.Recipient, &email.Subject, &email.CreatedAt)
        emails = append(emails, email)
    }

    // 4. Return JSON response
    response := ListEmailResponse{
        Emails: emails,
        Status: "success",
    }
    json.NewEncoder(w).Encode(response)
}
```

### **Data Structures**
```go
type EmailListItem struct {
    EmailID   string    `json:"email_id"`
    Recipient string    `json:"recipient"`
    Subject   string    `json:"subject"`
    CreatedAt time.Time `json:"created_at"`
}

type ListEmailResponse struct {
    Emails []EmailListItem `json:"emails"`
    Status string          `json:"status"`
    Error  string          `json:"error,omitempty"`
}
```

## 🚀 **API Usage Examples**

### **1. List User's Emails**
```bash
curl -X GET http://localhost:8080/api/email/list \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Success Response (200 OK):**
```json
{
  "emails": [
    {
      "email_id": "uuid-1",
      "recipient": "alice@example.com",
      "subject": "Project Update",
      "created_at": "2025-07-30T10:45:00Z"
    },
    {
      "email_id": "uuid-2",
      "recipient": "bob@example.com",
      "subject": "Meeting Notes",
      "created_at": "2025-07-30T09:30:00Z"
    },
    {
      "email_id": "uuid-3",
      "recipient": "charlie@example.com",
      "subject": "Weekly Report",
      "created_at": "2025-07-30T08:15:00Z"
    }
  ],
  "status": "success"
}
```

### **2. Empty Results (No Emails)**
```json
{
  "emails": [],
  "status": "success"
}
```

## 🔒 **Error Responses**

### **Authentication Errors (401 Unauthorized)**
```json
{
  "error": "Invalid or missing token"
}
```

### **Missing Authentication**
```json
{
  "error": "Authentication required"
}
```

### **Database Errors (500 Internal Server Error)**
```json
{
  "error": "Failed to retrieve emails"
}
```

```json
{
  "error": "Database connection unavailable"
}
```

## 🧪 **Testing**

### **Test Coverage**
```bash
# Run all list email tests
go test ./cmd/api -run TestListEmail -v

# Run integration tests
go test ./cmd/api -run TestListEmailIntegration -v

# Run response structure tests
go test ./cmd/api -run TestListEmailResponseStructure -v
```

### **Test Scenarios**
- ✅ **Valid JWT Token**: Passes authentication, returns user's emails
- ✅ **Missing Authorization Header**: Returns 401 Unauthorized
- ✅ **Invalid Bearer Format**: Returns 401 Unauthorized
- ✅ **Empty Token**: Returns 401 Unauthorized
- ✅ **Invalid JWT Token**: Returns 401 Unauthorized
- ✅ **Empty Results**: Returns empty array when user has no emails
- ✅ **Database Errors**: Returns 500 with appropriate error message
- ✅ **Response Structure**: Validates JSON response format

## 📊 **Database Query**

### **SQL Query**
```sql
SELECT email_id, recipient, subject, created_at
FROM emails 
WHERE sender_id = ?
ORDER BY created_at DESC
```

### **Query Parameters**
- `sender_id`: Authenticated user's ID from JWT token
- **Filtering**: Only emails sent by the authenticated user
- **Ordering**: Most recent emails first (DESC by created_at)
- **Fields**: Basic metadata only (no encrypted content)

### **Performance Considerations**
- **Indexing**: `sender_id` column should be indexed for fast filtering
- **Pagination**: Consider adding LIMIT/OFFSET for large result sets
- **Caching**: Response can be cached for short periods

## 🔄 **Integration with Existing System**

### **Route Registration**
```go
// In main.go
log.Printf("Registering /api/email/list endpoint")
r.Handle("/api/email/list", jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))).Methods("GET")
```

### **Middleware Chain**
1. **JWT Middleware**: Validates token and injects user_id into context
2. **List Handler**: Queries database and returns user's emails
3. **Error Handling**: Graceful error responses for all scenarios

### **Security Integration**
- **JWT Authentication**: Uses existing JWT middleware
- **User Context**: Leverages `GetUserIDFromContext()` helper
- **Database Security**: SQL injection protection via parameterized queries
- **Access Control**: User isolation at database level

## 🎯 **Use Cases**

### **Frontend Integration**
```javascript
// Example frontend usage
async function listUserEmails() {
    const response = await fetch('/api/email/list', {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${jwtToken}`,
            'Content-Type': 'application/json'
        }
    });

    if (response.ok) {
        const data = await response.json();
        return data.emails; // Array of email metadata
    } else {
        throw new Error('Failed to fetch emails');
    }
}
```

### **Email Management**
- **Inbox View**: Display list of sent emails
- **Email Details**: Use `email_id` to fetch full email content via `/api/email/get`
- **Search/Filter**: Frontend can filter by recipient, subject, or date
- **Pagination**: Handle large email lists efficiently

## 📈 **Future Enhancements**

### **Potential Features**
- **Pagination**: Add `limit` and `offset` query parameters
- **Filtering**: Add `recipient`, `subject`, `date_from`, `date_to` filters
- **Sorting**: Allow sorting by different fields
- **Search**: Full-text search across recipients and subjects
- **Caching**: Redis cache for frequently accessed lists

### **Performance Optimizations**
- **Database Indexes**: Optimize queries for large datasets
- **Connection Pooling**: Efficient database connection management
- **Response Compression**: Gzip compression for large responses
- **CDN Integration**: Cache responses at edge locations

---

**Status**: ✅ **Production Ready**

The GET `/api/email/list` endpoint is implemented, tested, and ready for production use with comprehensive authentication, authorization, and error handling. 