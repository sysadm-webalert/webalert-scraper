# WebAlert scraper service
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=sysadm-webalert_webalert-scraper&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=sysadm-webalert_webalert-scraper)
## Overview
A microservice that scrapes monitored websites to collect performance metrics and status.

## Features
- **Site Scraping**:  Gathers metrics such as status codes, response times, page load duration, and page sizes.
- **Metric Collector**: Sends collected metrics to the backend for further processing and analysis.

## Installation

### Prerequisites
- Go ^1.23.2 (for building from source)

### Local Build
1. Install required packages
   ```bash
   go mod download
   ```
2. Build the service
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -o webalert-webscraper
   ```

### Docker Build
1. Build the image
   ```bash
   docker compose --profile dev build 
   ```
2. Run the image
   ```bash
   docker compose --profile dev  up -d
   ```

## Configuration
1. The following environment variables must be set with the correspondent data:
    ```bash
    WEBALERT_BACKEND_USER: "application_user"
    WEBALERT_BACKEND_PASSWORD: "application_user_password"
    WEBALERT_BACKEND_LOGIN_URL: "http://webalert-backend-dev"
   ```

## Contributing
We welcome contributions! Please follow these steps:
1. Fork the repository.
2. Create a feature branch.
3. Commit your changes.
4. Open a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support
For issues or feature requests, please open an issue in the [GitHub repository](https://github.com/sysadm-webalert/webalert-scraper/issues).

---
**WebAlert Agent** © 2024