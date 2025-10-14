# BuilDefect - monolithic web application for centralized management of defects at construction sites.

The system provides a full cycle of work: from defect registration and assignee allocation to status monitoring and report generation for management.

Who the system is for:
1) Engineers (defect registration, information updates)
2) Managers (task assignment, deadline monitoring, report generation)
3) Executives and Clients (progress tracking and report viewing)


### ER-diagram

![alt text](https://github.com/Quasar777/buildefect/blob/main/buisness%20analytics/BuilDefect_ER.drawio.png?raw=true)


### How to start 

Before running the project, make sure you have the following software installed on your machine:

- [Go](https://golang.org/doc/install) (version 1.20 or higher recommended)
- [Git](https://git-scm.com/downloads)
- Optional: [Postman](https://www.postman.com/) or any other API testing tool for testing endpoints

#### Backend:

1) clone this repostirory by using a command

```bash
git clone https://github.com/Quasar777/buildefect.git
```

2) navigate to the backend directory

```bash
cd buildefect/backend
```

3) run the server

```bash
go run cmd/api/main.go
```

The server will start, and by default, it should be accessible at http://localhost:8080 (adjust the port if configured differently).


#### Frontend

Before running the frontend, make sure you have the following installed on your machine:

* [Node.js](https://nodejs.org/) (version 18 or higher recommended)
* [npm](https://www.npmjs.com/) or [Yarn](https://yarnpkg.com/) package manager
* Optional: [Git](https://git-scm.com/downloads) to clone the repository


1. Navigate to the frontend directory:

```bash
cd buildefect/frontend
```

2. Install dependencies:

```bash
npm install
# or, if you use Yarn:
# yarn install
```

3. Start the development server:

```bash
npm run start
# or, if you use Yarn:
# yarn start
```

The frontend server will start, usually accessible at [http://localhost:3000](http://localhost:3000).
Make sure the backend is running (default at [http://localhost:8080](http://localhost:8080)) so that API requests work correctly.

