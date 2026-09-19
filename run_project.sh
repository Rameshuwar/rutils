#!/bin/bash

# Main Menu
show_main_menu() {
    echo "=============================="
    echo "          Main Menu           "
    echo "=============================="
    echo "1) run project"
    echo "2) test check"
    echo "3) exit"
    echo "=============================="
    read -p "Enter your choice (1-3): " main_choice

    case $main_choice in
        1)
            show_run_project_menu
            ;;
        2)
            show_test_check_menu
            ;;
        3)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo "Invalid option. Please try again."
            echo ""
            show_main_menu
            ;;
    esac
}

# Run Project Menu
show_run_project_menu() {
    echo ""
    echo "=============================="
    echo "       Run Project Menu       "
    echo "=============================="
    echo "1) to run full project"
    echo "2) backend(swagger)"
    echo "3) back to main menu"
    echo "=============================="
    read -p "Enter your choice (1-3): " run_choice

    case $run_choice in
        1)
            echo ""
            echo "Starting full project (Frontend & Backend)..."
            if [ -f "./dev.sh" ]; then
                ./dev.sh
            else
                echo "Error: dev.sh not found."
            fi
            ;;
        2)
            echo ""
            echo "Starting Backend with Swagger..."
            echo "Backend API and Swagger UI will be available at:"
            echo " 🌐 http://localhost:8080/swagger/"
            echo "Press [Ctrl+C] to stop."
            go run ./cmd/server/main.go
            ;;
        3)
            echo ""
            show_main_menu
            ;;
        *)
            echo "Invalid option. Please try again."
            show_run_project_menu
            ;;
    esac
}

# Test Check Menu
show_test_check_menu() {
    echo ""
    echo "=============================="
    echo "       Test Check Menu        "
    echo "=============================="
    echo "1) run to the frontend test"
    echo "2) backend test"
    echo "3) back to main menu"
    echo "=============================="
    read -p "Enter your choice (1-3): " test_choice

    case $test_choice in
        1)
            echo ""
            echo "Running frontend tests..."
            if [ -f "./persona-tests/run-ui-tests.sh" ]; then
                (cd persona-tests && ./run-ui-tests.sh)
            else
                echo "Error: ./persona-tests/run-ui-tests.sh not found."
            fi
            ;;
        2)
            echo ""
            echo "Running backend tests..."
            if [ -f "./persona-tests/run-api-tests.sh" ]; then
                (cd persona-tests && ./run-api-tests.sh)
            else
                echo "Error: ./persona-tests/run-api-tests.sh not found."
            fi
            ;;
        3)
            echo ""
            show_main_menu
            ;;
        *)
            echo "Invalid option. Please try again."
            show_test_check_menu
            ;;
    esac
}

# Start the script by showing the main menu
show_main_menu
