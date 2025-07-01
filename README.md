# SEA Catering Backend

> **Healthy Meals, Anytime, Anywhere** - A comprehensive meal catering management system built with Go and modern technologies.

## 🍽️ Overview

SEA Catering Backend is a robust REST API service that powers a meal subscription platform. The system enables users to subscribe to customizable meal plans, manage deliveries, and provides comprehensive admin controls for business management.

## ✨ Features

### 🔐 Authentication & Authorization
- **JWT-based authentication** with secure token management
- **Email verification** with OTP system
- **Password reset** functionality
- **Role-based access control** (User/Admin)
- **Multi-factor authentication** support

### 🍛 Meal Plan Management
- **Flexible meal plans** (Diet, Protein, Royal)
- **Customizable meal types** (Breakfast, Lunch, Dinner)
- **Dynamic pricing** based on selections
- **Image upload** support with S3 integration
- **Search and filtering** capabilities

### 📅 Subscription System
- **Flexible delivery scheduling** (Monday-Sunday)
- **Subscription management** (Active, Paused, Cancelled)
- **Automatic pause/resume** functionality
- **Allergy and dietary preferences** support
- **Subscription reactivation** for cancelled plans

### 💬 Customer Reviews
- **Testimonial submission** with rating system
- **Admin moderation** (Approve/Reject)
- **Public testimonial display** for approved reviews

### 📊 Admin Dashboard
- **Comprehensive analytics** and reporting
- **User management** with detailed profiles
- **Subscription oversight** and control
- **Revenue tracking** and growth metrics
- **Testimonial moderation** tools

### 🔧 Technical Features
- **Structured logging** with request tracing
- **Rate limiting** and security middleware
- **Email notifications** with HTML templates
- **File upload** to AWS S3
- **Database migrations** with versioning
- **Redis caching** for performance
- **Payment integration** ready (Midtrans)

## 🏗️ Architecture

```
sea-catering-backend/
├── cmd/app/                    # Application entrypoint
├── database/
│   ├── migrations/            # Database schema migrations
│   ├── seeds/                 # Sample data
│   └── postgres/              # Database configuration
├── internal/
│   ├── api/                   # API modules
│   │   ├── admin/            # Admin management
│   │   ├── auth/             # Authentication
│   │   ├── meal_plans/       # Meal plan operations
│   │   ├── subscriptions/    # Subscription management
│   │   └── testimonials/     # Customer reviews
│   ├── config/               # Application configuration
│   ├── entity/               # Domain models
│   └── middleware/           # HTTP middleware
├── pkg/                      # Shared packages
│   ├── bcrypt/              # Password hashing
│   ├── email/               # Email service
│   ├── jwt/                 # JWT utilities
│   ├── logger/              # Structured logging
│   ├── redis/               # Redis client
│   ├── response/            # HTTP responses
│   ├── s3/                  # AWS S3 integration
│   └── utils/               # Utility functions
└── tools/migration/         # Migration tool
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 14+**
- **Redis 6+**
- **AWS S3** (for file storage)
- **SMTP Server** (for emails)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-org/sea-catering-backend.git
cd sea-catering-backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Run database migrations**
```bash
go run tools/migration/main.go -command=up
```

5. **Seed the database** (optional)
```bash
psql -d sea_catering -f database/seeds/001_admin_users.sql
psql -d sea_catering -f database/seeds/002_meal_plans.sql
psql -d sea_catering -f database/seeds/003_sample_testimonials.sql
```

6. **Start the server**
```bash
go run cmd/app/main.go
```

The API will be available at `http://localhost:8080`

## 🐳 Docker Deployment

### Using Docker Compose

1. **Copy Docker environment**
```bash
cp .env.docker .env
```

2. **Start services**
```bash
docker-compose up -d
```

3. **Run migrations**
```bash
docker-compose exec backend go run tools/migration/main.go -command=up
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Server port | `8080` |
| `APP_ENV` | Environment | `development` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `sea_catering` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `JWT_SECRET` | JWT signing key | - |
| `SMTP_HOST` | SMTP server | `smtp.gmail.com` |
| `SMTP_USERNAME` | SMTP username | - |
| `SMTP_PASSWORD` | SMTP password | - |
| `AWS_BUCKET_NAME` | S3 bucket name | - |

See `.env.example` for complete configuration options.

## 📡 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/forgot-password` - Password reset request
- `POST /api/v1/auth/reset-password` - Password reset
- `POST /api/v1/auth/send-otp` - Send OTP verification
- `POST /api/v1/auth/verify-otp` - Verify OTP

### User Management
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update profile
- `POST /api/v1/user/change-password` - Change password
- `POST /api/v1/user/profile/image` - Upload profile image

### Meal Plans
- `GET /api/v1/meal-plans` - List all meal plans
- `GET /api/v1/meal-plans/active` - Get active meal plans
- `GET /api/v1/meal-plans/search` - Search meal plans
- `GET /api/v1/meal-plans/{id}` - Get meal plan details
- `GET /api/v1/meal-plans/popular` - Get popular meal plans

### Subscriptions
- `POST /api/v1/subscriptions` - Create subscription
- `GET /api/v1/subscriptions/my` - Get user subscriptions
- `GET /api/v1/subscriptions/{id}` - Get subscription details
- `PUT /api/v1/subscriptions/{id}` - Update subscription
- `PUT /api/v1/subscriptions/{id}/pause` - Pause subscription
- `PUT /api/v1/subscriptions/{id}/resume` - Resume subscription
- `PUT /api/v1/subscriptions/{id}/reactivate` - Reactivate cancelled subscription
- `DELETE /api/v1/subscriptions/{id}` - Cancel subscription

### Testimonials
- `POST /api/v1/testimonials` - Submit testimonial
- `GET /api/v1/testimonials` - Get approved testimonials

### Admin Endpoints
- `POST /api/v1/admin/login` - Admin login
- `GET /api/v1/admin/dashboard` - Dashboard statistics
- `POST /api/v1/admin/dashboard/filter` - Filtered statistics

#### Admin - User Management
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/users/{id}` - Get user details
- `PUT /api/v1/admin/users/{id}/status` - Update user status
- `DELETE /api/v1/admin/users/{id}` - Delete user

#### Admin - Meal Plans
- `POST /api/v1/meal-plans/admin` - Create meal plan
- `PUT /api/v1/meal-plans/admin/{id}` - Update meal plan
- `DELETE /api/v1/meal-plans/admin/{id}` - Delete meal plan
- `PATCH /api/v1/meal-plans/admin/{id}/activate` - Activate meal plan
- `PATCH /api/v1/meal-plans/admin/{id}/deactivate` - Deactivate meal plan

#### Admin - Subscriptions
- `GET /api/v1/subscriptions/admin/search` - Search subscriptions
- `PUT /api/v1/subscriptions/admin/{id}/force-cancel` - Force cancel subscription

#### Admin - Testimonials
- `GET /api/v1/testimonials/admin/all` - Get all testimonials
- `PUT /api/v1/testimonials/admin/{id}/approve` - Approve testimonial
- `PUT /api/v1/testimonials/admin/{id}/reject` - Reject testimonial
- `DELETE /api/v1/testimonials/admin/{id}` - Delete testimonial

## 🗄️ Database Schema

### Core Tables
- **users** - User accounts and authentication
- **admin_users** - Administrative accounts
- **meal_plans** - Available meal plans
- **subscriptions** - User subscriptions
- **testimonials** - Customer reviews
- **subscription_audit** - Subscription change history

### Key Relationships
```sql
users (1) ←→ (n) subscriptions
meal_plans (1) ←→ (n) subscriptions
subscriptions (1) ←→ (n) subscription_audit
```

## 📝 Logging

The application uses structured logging with multiple output formats:

### Log Levels
- **DEBUG** - Detailed diagnostic information
- **INFO** - General operational messages
- **WARN** - Warning conditions
- **ERROR** - Error conditions
- **FATAL** - Critical errors causing shutdown

### Log Outputs
- **Console** - Formatted output for development
- **File** - JSON format for production (`./logs/app.log`)
- **Both** - Combined output

## 🔒 Security Features

- **Password hashing** with bcrypt
- **JWT token** authentication
- **Rate limiting** per IP
- **CORS** protection
- **Input validation** and sanitization
- **SQL injection** protection
- **XSS** prevention
- **Request ID** tracing

## 🚀 Performance

- **Redis caching** for frequently accessed data
- **Database indexing** for optimal queries
- **Connection pooling** for database connections
- **Gzip compression** for HTTP responses
- **Static file** serving with CDN support

## 📈 Monitoring

### Health Checks
- `GET /health` - Application health status
- Database connectivity check
- Redis connectivity check

### Metrics
- Request duration and count
- Database query performance
- Cache hit/miss rates
- Error rates by endpoint

## 🔧 Migration Management

### Migration Commands
```bash
# Create new migration
go run tools/migration/main.go -command=create -name="add_user_preferences"

# Apply migrations
go run tools/migration/main.go -command=up

# Rollback migrations
go run tools/migration/main.go -command=down -steps=1

# Check current version
go run tools/migration/main.go -command=version

# Force specific version
go run tools/migration/main.go -command=force -version=5
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines
- Follow Go conventions and best practices
- Write comprehensive tests for new features
- Update documentation for API changes
- Use meaningful commit messages
- Ensure all tests pass before submitting

## 🙋‍♂️ Support

### Getting Help
- **Documentation** - Check this README and inline code comments
- **Issues** - Create an issue for bug reports or feature requests
- **Discussions** - Use GitHub Discussions for questions

## 🔄 Changelog

### v1.0.0 (Current)
- ✅ User authentication and authorization
- ✅ Meal plan management
- ✅ Subscription system
- ✅ Admin dashboard
- ✅ Email notifications
- ✅ File upload integration
- ✅ Testimonial system
