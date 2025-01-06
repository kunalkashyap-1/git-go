# Git-Go: A Simple Git Implementation

Git-Go is a lightweight, custom-built implementation of Git commands in go. It helps you understand the core mechanisms behind Git operations by implementing essential commands like `init`, `cat-file`, `hash-object`, `ls-tree`, `write-tree`, and `commit-tree`. This project serves as a hands-on learning tool for understanding how Git manages files, commits, and repositories under the hood. I also attempted to use the **Strategy Pattern** for the first time to organize the command execution logic in a clean, maintainable way.

## Features

- **`init`**: Initializes a new git repository, creating necessary directories.
- **`cat-file`**: Views the contents of an object stored in the git object.
- **`hash-object`**: Computes the hash of a file, stores it in the object, and compresses the data.
- **`ls-tree`**: Lists the contents of a tree object in a git repository.
- **`write-tree`**: Creates a tree object from the current directory structure.
- **`commit-tree`**: Creates a commit object with a reference to a tree and parent commit, along with a commit message.

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/kunalkashyap-1/git-go.git
    cd git-go
    ```
2. Run Git-Go:

    ```bash
    ./your_program.sh <command> [<args>...]
    ```

## Usage

Each command in Git-Go works similarly to its Git counterpart. Below is an example of how you can use these commands.

- **`init`**: Initializes a new repository.

    ```bash
    ./your_program.sh init
    ```

- **`cat-file`**: Shows the contents of a git object.

    ```bash
    ./your_program.sh cat-file -p <commit-sha>
    ```

- **`hash-object`**: Computes the hash of a file and stores it.

    ```bash
    ./your_program.sh hash-object -w <file-path>
    ```

- **`ls-tree`**: Lists the contents of a tree object.

    ```bash
    ./your_program.sh ls-tree --name-only <commit-sha> 
    ```

- **`write-tree`**: Creates a tree object for the current directory.

    ```bash
    ./your_program.sh write-tree
    ```

- **`commit-tree`**: Creates a commit object for a given tree and parent commit.

    ```bash
    ./your_program.sh commit-tree <tree-sha> -p <parent-sha> -m "commit message"
    ```


## Inspiration

This project was made following the blueprint provided by **Codecrafters**. The goal is to offer a deeper understanding of Git by replicating its core functionality from scratch. This approach helps illuminate how Git stores objects, manages commits, and handles repository structures. By working with these basic Git operations, you'll gain a more profound understanding of how Git works internally, which can aid in both debugging complex issues and developing new tools.

In addition, I experimented with using the **Strategy Pattern** for the first time to organize the various git commands in a structured way, making the codebase more modular and easier to extend.

