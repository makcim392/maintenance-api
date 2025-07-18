# Maintenance API - Comprehensive Improvement Plan

## Overview
This document outlines the complete refactoring plan for the maintenance API project to address feedback issues and prepare it for public job portfolio use.

## Issues to Address
1. **Handler Organization**: Mixed business logic, SQL, and HTTP responsibilities
2. **Missing Validations**: Empty username/password, task summary without minimum length
3. **Makefile Clarity**: Unclear test commands and usage
4. **Company References**: Remove all "swordhealth" references
5. **Testing**: Enhance test coverage and clarity

## Phase 1: Architecture Refactoring (Clean Architecture)

### New Directory Structure
```
internal/
├── repository/         # Data access layer
│   ├── user_repository.go
│   ├── task_repository.go
│   └── interfaces.go
├── service/           # Business logic layer
│   ├── auth_service.go
│   ├── task_service.go
│   └── interfaces.go
├── handler/           # HTTP handlers (thin layer)
│   ├── auth_handler.go
│   ├── task_handler.go
│   └── dto/          # Data transfer objects
│       ├── auth_dto.go
│       └── task_dto.go
├── validator/         # Request validation
│   ├── validators.go
│   └── errors.go
├── models/            # Domain models (existing)
├── middleware/        # HTTP middleware (existing)
└── auth/              # JWT handling (existing)
```

### Repository Layer Responsibilities
- All database operations (CRUD)
- Database connection management
- Query building and execution
- Transaction handling

### Service Layer Responsibilities
- Business logic implementation
- Authorization rules
- Data transformation
- Validation coordination

### Handler Layer Responsibilities
- HTTP request/response handling
- Request parsing and validation
- Error response formatting
- Delegation to service layer

## Phase 2: Enhanced Validation

### Validation Rules
| Field | Rules | Error Message |
|-------|--------|---------------|
| Username | Required, 3-50 chars, alphanumeric | "Username must be 3-50 alphanumeric characters" |
| Password | Required, 8+ chars, 1 uppercase, 1 lowercase, 1 number | "Password must be 8+ chars with uppercase, lowercase, and number" |
| Task Summary | Required, 10-2500 chars | "Summary must be 10-2500 characters" |
| Task Date | Required, not future date | "Task date cannot be in the future" |
| Role | Required, must be "technician" or "manager" | "Role must be either 'technician' or 'manager'" |

### Implementation
- Use `go-playground/validator` package
- Custom validators for business rules
- Centralized error handling
- Consistent error response format

## Phase 3: Makefile & Documentation

### New Makefile Structure
```makefile
# Development
make dev          # Run with hot reload using Air
make build        # Build the application
make run          # Run the application

# Testing
make test         # Run unit tests only
make test-integration  # Run integration tests only
make test-all     # Run all tests
make test-cover   # Run tests with coverage
make test-cover-html   # Generate HTML coverage report

# Code Quality
make fmt          # Format code
make lint         # Run linter
make tidy         # Tidy dependencies

# Database
make db-up        # Start database containers
make db-down      # Stop database containers
make db-reset     # Reset database

# Help
make help         # Show all available commands
```

### Documentation Updates
- Clear testing instructions in README
- API documentation with examples
- Development setup guide
- Contributing guidelines

## Phase 4: Remove Company References

### Files to Update
- `go.mod` - Module name
- All import paths in `.go` files
- Docker configurations
- Documentation files

### New Module Name
`github.com/makcim392/maintenance-api`

## Phase 5: Enhanced Testing

### Test Structure
```
tests/
├── unit/                    # Unit tests
│   ├── repository/
│   ├── service/
│   └── validator/
├── integration/             # Integration tests (existing)
└── fixtures/               # Test data
```

### Test Coverage Goals
- Repository layer: 90%+ coverage
- Service layer: 85%+ coverage
- Handler layer: 80%+ coverage
- Overall: 85%+ coverage

### Test Types
1. **Unit Tests**: Individual functions/methods
2. **Repository Tests**: Database operations with test DB
3. **Service Tests**: Business logic with mocked repositories
4. **Handler Tests**: HTTP handlers with mocked services
5. **Integration Tests**: End-to-end API testing

## Implementation Timeline

### Week 1: Foundation
- [ ] Phase 4: Remove company references
- [ ] Phase 3: Update Makefile and documentation
- [ ] Set up new directory structure

### Week 2: Core Refactoring
- [ ] Phase 1: Implement repository layer
- [ ] Phase 1: Implement service layer
- [ ] Phase 1: Refactor handler layer

### Week 3: Validation & Testing
- [ ] Phase 2: Implement validation layer
- [ ] Phase 5: Write comprehensive tests
- [ ] Update existing tests

### Week 4: Polish & Documentation
- [ ] Final testing and bug fixes
- [ ] Complete documentation
- [ ] Performance optimization
- [ ] Security review

## Technical Decisions

### Dependencies to Add
- `github.com/go-playground/validator/v10` - Request validation
- `github.com/stretchr/testify` - Testing utilities (already present)
- `github.com/DATA-DOG/go-sqlmock` - Database mocking (already present)

### Dependencies to Review
- Remove any unused dependencies
- Update to latest stable versions

### Error Handling Strategy
- Consistent error response format
- Proper HTTP status codes
- Detailed error messages for debugging
- Structured logging

### Security Considerations
- Input sanitization
- SQL injection prevention (using prepared statements)
- Password hashing verification
- Rate limiting consideration

## Success Criteria

### Code Quality Metrics
- [ ] All handlers under 100 lines of code
- [ ] No SQL queries in handlers
- [ ] 100% validation coverage for all endpoints
- [ ] Zero company references
- [ ] 85%+ test coverage

### Documentation Completeness
- [ ] Clear README with setup instructions
- [ ] API documentation with examples
- [ ] Testing guide
- [ ] Architecture explanation

### Developer Experience
- [ ] Single command setup
- [ ] Clear error messages
- [ ] Consistent code style
- [ ] Helpful Makefile commands

## Risk Mitigation

### Potential Issues
1. **Breaking Changes**: Maintain backward compatibility during refactoring
2. **Test Failures**: Keep existing tests passing while adding new ones
3. **Performance**: Monitor for any performance regressions
4. **Complexity**: Ensure new architecture doesn't over-engineer simple features

### Rollback Plan
- Keep git commits small and focused
- Tag stable versions before major changes
- Maintain working version on main branch

## Next Steps
1. Review and approve this plan
2. Toggle to Act mode to begin implementation
3. Start with Phase 4 (company reference removal) as it's the simplest
4. Proceed systematically through each phase
