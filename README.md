# NcRegistry

The NcRegistry Go binary is a powerful command-line tool designed to streamline interaction with Docker registries. This tool is written in Go and provides users with an interactive prompt to select registries, view repositories, and manage Docker images and tags.

Key Features:

* **Registry Selection:** Connects to multiple Docker registries, allowing easy selection and switching.
* **Repository Browsing:** Fetches and displays the list of repositories in the selected registry.
* **Tag Management:** Fetches Docker tags and displays metadata, including creation dates and size. Enables actions such as pulling images and deleting manifests directly from the command-line interface.
* **Ease of Use:** Intuitive prompts guide the user through each action, simplifying complex operations.

This tool is a must-have for developers and administrators who frequently work with Docker and desire a more efficient and manageable way to interact with Docker registries.

<br>

## Installation via Homebrew (MacOS/Linux - x86_64/arm64)
```
brew install stenstromen/tap/ncregistry
```

## Download and Run Binary
* For **MacOS** and **Linux**: Checkout and download the latest binary from [Releases page](https://github.com/Stenstromen/ncregistry/releases/latest/)
* For **Windows**: Build the binary yourself.

## Build and Run Binary
```
go build
./ncregistry
```

## Example Usage 
<div style="background-color: black">
~$ ncregistry<br>
... (screen clears)<br><br>
<strong>Use ↑/↓ to move and Enter to select</strong><br>
<span style="color: blue">?</span> Main Menu:<br>
&nbsp;&nbsp;👉 <strong style="color:cyan">Connect to registry</strong><br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; <span style="color:cyan">Add registry</span><br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; <span style="color:cyan">Remove registry</span><br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; <span style="color:cyan">Exit</span><br>
</div>

## Limitations
* Currently only supports (tested) with the official Docker registry server (registry).
* Only supports Docker v2 API.
* Only supports basic authentication (username/password).
* Only supports HTTPS connections (no HTTP).