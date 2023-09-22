---
name: Backend Service
interfaces:
  API:
    description: RESTful API
    tags:
      link: https://www.openapis.org/
dependencies:
  ToDB:
    depends_on: BE.DB.SQL
    description: persists data to
tags:
  link: https://github.com/libsql/libsql
---
