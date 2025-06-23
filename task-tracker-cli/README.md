# Task Tracker CLI

A simple command-line application built in Golang to manage tasks. This application allows you to add, update, delete, list, and mark the status of tasks. Tasks are stored in a JSON file for persistence.

Roadmap Project Challenge: https://roadmap.sh/projects/task-tracker

## Features

- **Add Task**: Create a new task with a description.
- **Update Task**: Modify the description or status of an existing task.
- **Delete Task**: Remove a task from the list.
- **Mark Task Status**: Mark a task as in progress or done.
- **List Tasks**: View all tasks along with their description and status.

## Installation

To install and run the Task Tracker CLI, ensure you have Golang installed, then follow these steps:

```bash
# Clone the repository
git clone https://github.com/abneed/task-tracker-cli.git

# Navigate to the project directory
cd task-tracker-cli

# Build the application
go build -o task-cli

# Run the application
./task-cli
```

## Usage
Here’s how you can use the various commands:

### Add a Task
```bash
./task-cli add "Your task description"
```

### Update a Task
```bash
./task-cli update <task_id> "Updated task description"
```

### Delete a Task
```bash
./task-cli delete <task_id>
```

### Marking a task as "in-progress" or "done"
```bash
./task-cli mark-in-progress <task_id>
```

```bash
./task-cli mark-done <task_id>
```

### Listing all tasks
```bash
./task-cli list
```

## Task Storage

Tasks are stored in a JSON file located in the project directory. This allows for easy persistence and manipulation of task data.

In case that the JSON file doesn't exists, the application will generate that directory and the JSON file storage automatically `db/tasks.json`.

## Contributing
Contributions are welcome! Please fork the repository and submit a pull request for review.

## License
This project is licensed under the MIT License - see the [LICENSE](https://github.com/abneed/task-tracker-cli/blob/master/LICENSE) file for details.