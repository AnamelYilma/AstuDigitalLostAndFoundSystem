FINAL PROJECT - 1st & 2nd Year

Project Title: ASTU Digital Lost & Found System

Background & Problem Statement
Students at ASTU frequently lose personal belongings (ID cards, calculators, USB drives, lab coats, books, etc.).
This system provides a centralized digital platform to report, search, and manage lost/found items.

Project Objective
Build a secure platform to:
- Report lost items
- Report found items
- Search items
- Submit claims for found items
- Track approval workflow and status

System Roles
1. Student
- Register and login with Student ID + password
- Report lost/found item
- Search items (with or without login)
- Open item details and claim found items (login required)
- See own report status (pending/approved/rejected)

2. Admin
- Login from separate Admin Login page
- Approve/reject/remove item posts
- Approve/reject claim requests
- View user contact info for coordination
- View dashboard statistics

Campus Location List (ASTU)
- Library
- Cafe
- Class
- Lap
- Dorm
- On Road
- Tolest
- Shower
- Anphe
- Launch
- Park
- Hale.Birroe
- Other

Core Functionalities
- User registration (Name, Student ID, Phone, Password)
- Student login by Student ID + Password
- Separate admin login page
- Report lost/found with image upload
- Location selection from ASTU location list
- Search and filter by type, category, location, and color
- Public item search/list/detail available without login
- Claim submission requires login
- Admin approval workflow for item posts
- Claim approval/rejection workflow
- User dashboard showing approval + claim status

Workflow Rules
1. User submits item report -> status starts as pending
2. Admin approves/rejects/removes item report
3. Only approved items appear in public search
4. User can open large item detail without login
5. To apply/claim from detail page, user must login
6. User remains logged in until Logout is clicked

Security Requirements
- Password hashing
- Role-based access control
- Protected admin routes
- Input validation
- Secure upload handling
- Prevent unauthorized moderation/claim actions

Quick Run
1. Start PostgreSQL and ensure DB config in pkg/database/db.go is correct
2. Run: go run main.go
3. Open: http://localhost:8080
