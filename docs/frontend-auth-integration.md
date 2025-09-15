# Frontend Authentication Integration Guide

## Overview

This guide provides essential information for integrating authentication with the Strive API backend from a frontend application.

## Base URL

```
http://localhost:8080
```

## CORS Configuration

The API is configured to accept requests from the following origins:
- `http://localhost:3000` (React default)
- `http://localhost:3001` (React alternative)
- `http://localhost:4200` (Angular default)
- `http://127.0.0.1:3000`
- `http://127.0.0.1:3001`
- `http://127.0.0.1:4200`

## Authentication Flow

The API uses JWT (JSON Web Token) based authentication with access and refresh tokens.

### Token Types

- **Access Token**: Short-lived (15 minutes), used for API requests
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens
- **Token Type**: Bearer

## API Endpoints

### 1. User Registration

**Endpoint**: `POST /api/v1/auth/register`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (201 Created):
```json
{
  "message": "User registered successfully",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Response** (400 Bad Request):
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data"
  }
}
```

### 2. User Login

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Error Response** (401 Unauthorized):
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  }
}
```

### 3. Refresh Token

**Endpoint**: `POST /api/v1/auth/refresh`

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Error Response** (401 Unauthorized):
```json
{
  "error": {
    "code": "INVALID_REFRESH_TOKEN",
    "message": "Invalid or expired refresh token"
  }
}
```

## Protected Endpoints

All endpoints under `/api/v1/user/` require authentication.

### Authorization Header

Include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Example Protected Request

```javascript
fetch('/api/v1/user/profile', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + accessToken,
    'Content-Type': 'application/json'
  }
})
```

## Error Handling

### Common Error Codes

- `UNAUTHORIZED`: Missing or invalid authorization header
- `INVALID_TOKEN`: Token is malformed or expired
- `INVALID_CREDENTIALS`: Wrong email/password combination
- `INVALID_REFRESH_TOKEN`: Invalid or expired refresh token
- `VALIDATION_ERROR`: Invalid input data format
- `REGISTRATION_FAILED`: User already exists or registration failed

### Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

## Frontend Implementation Examples

### JavaScript/TypeScript

```typescript
class AuthService {
  private baseURL = 'http://localhost:8080';
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  async register(email: string, password: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error.message);
    }
  }

  async login(email: string, password: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error.message);
    }

    const data = await response.json();
    this.accessToken = data.access_token;
    this.refreshToken = data.refresh_token;
    
    // Store tokens in localStorage or secure storage
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
  }

  async makeAuthenticatedRequest(url: string, options: RequestInit = {}): Promise<Response> {
    const token = this.accessToken || localStorage.getItem('access_token');
    
    if (!token) {
      throw new Error('No access token available');
    }

    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers,
    };

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (response.status === 401) {
      // Token expired, try to refresh
      const refreshed = await this.refreshAccessToken();
      if (refreshed) {
        // Retry the original request with new token
        const newToken = this.accessToken || localStorage.getItem('access_token');
        const newHeaders = {
          ...headers,
          'Authorization': `Bearer ${newToken}`,
        };
        
        return fetch(url, {
          ...options,
          headers: newHeaders,
        });
      } else {
        // Refresh failed, redirect to login
        this.handleTokenExpired();
      }
    }

    return response;
  }

  async refreshAccessToken(): Promise<boolean> {
    const refreshToken = this.refreshToken || localStorage.getItem('refresh_token');
    
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        return false;
      }

      const data = await response.json();
      this.accessToken = data.access_token;
      this.refreshToken = data.refresh_token;
      
      // Update stored tokens
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      
      return true;
    } catch (error) {
      console.error('Failed to refresh token:', error);
      return false;
    }
  }

  private handleTokenExpired(): void {
    // Clear tokens and redirect to login
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    // Redirect to login page
  }

  logout(): void {
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }
}
```

### React Hook Example

```typescript
import { useState, useEffect } from 'react';

export const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    setIsAuthenticated(!!token);
    setLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    const authService = new AuthService();
    await authService.login(email, password);
    setIsAuthenticated(true);
  };

  const logout = () => {
    const authService = new AuthService();
    authService.logout();
    setIsAuthenticated(false);
  };

  return { isAuthenticated, loading, login, logout };
};
```

## Security Considerations

1. **Token Storage**: Store tokens securely (consider using httpOnly cookies for production)
2. **HTTPS**: Always use HTTPS in production
3. **Token Expiration**: Implement automatic token refresh or re-authentication
4. **CORS**: Configure CORS properly for your frontend domain
5. **Input Validation**: Validate all user inputs on the frontend before sending to API
6. **Automatic Refresh**: The provided AuthService automatically handles token refresh on 401 responses

## Validation Rules

### Email
- Must be a valid email format
- Required field

### Password
- Minimum 8 characters
- Required field

## Testing

Use the health endpoint to verify API connectivity:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok"
}
```

## Swagger Documentation

Interactive API documentation is available at:
```
http://localhost:8080/swagger/
```
