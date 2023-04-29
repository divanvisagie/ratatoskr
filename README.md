# Ratatoskr
Ratatoskr is a Telegram bot designed to help you deal with web links that you would usually paste into the "Saved Messages" chat. It is named after the squirrel Ratatoskr from Norse mythology, who would run up and down the world tree Yggdrasil to deliver messages between the eagle at the top and the serpent at the bottom.

We don't have any mythological trees, but the idea is that when you paste a link here, Ratatoskr will take the appropriate action, for example for basic web pages it 
will read and summarise the link as well as save it for later reading in a notion database. For Youtube links it will process the video audio using OpenAI whisper, return a summary of the video and save the video and the summary to a notion database. For links to PDFs it will save the PDF to a notion database. 

![Ratatoskr logo](./docs/logo-256.png)

## Architecture
The architecture is constructed by two main concepts, layers and capabilities. Layers are responsible for intercepting RequestMessages and processing them in some way. For example the MemoryLayer will store the messages that pass through it but also enrich RequestMessages with a context for the conversation history for that user.

Capabilities are responsible for executing a specific task, they also contain a `Check` function that determines if the capability should be executed for a given RequestMessage. For example the LinkProcessorCapability will check if the RequestMessage contains a link and if so it will process it.

    
### User Flow
```mermaid
sequenceDiagram

    actor U as user
    participant TB as Telegram Bot

    participant H as Handler
    box Layers
    participant ML as Memory(Layer)
    participant CS as CapabilitySelector(Layer)
    end

    box Capabilities
    participant G as GPT(Capability)
    participant C as LinkProcessor(Capability)
    end

    U->>TB: User sends message
    TB->>H: Handle message
    H->>H: Convert to RequestMessage
    H->>ML: Forward RequestMessage
    ML->>CS: Forward RequestMessage

    loop Every Capability
    CS->>G: Run Check RequestMessage
    G-->>CS: Returns Response
    end

    alt Is a question
    CS->>G: Execute RequestMessage
    else Is a link
    CS->>C: Execute RequestMessage
    end

    CS-->>ML: Returns Response
    ML-->>H: Returns Response
    H-->>TB: Returns Response
    TB->>U: Returns Response
```


## Setup
```sh
go install github.com/cosmtrek/air@latest
```

**Run**
```
air
```