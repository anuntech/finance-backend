# Finance Backend

This project is a backend service for managing financial accounts, banks, expenses, recipes, and workspaces. It is built using Go and MongoDB.

## Table of Contents

- [Getting Started](#getting-started)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

These instructions will help you set up and run the project on your local machine for development and testing purposes.

## Prerequisites

- Go 1.18 or later
- MongoDB
- A running instance of MongoDB

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/anuntech/finance-backend.git
   cd finance-backend
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set up your MongoDB connection string and database name in the environment variables or configuration file.

## Usage

To run the project, use the following command:

use `gow run .` to execute the project