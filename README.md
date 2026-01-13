## System Architecture (Mermaid)

```mermaid
graph TD
    subgraph Sources ["External Sources"]
        LogFile[File Tail / Syslog]
        SDK[User Apps SDK]
        Cloud[CloudWatch / Datadog Webhooks]
    end

    subgraph IngestionService ["Ingestion Service (Go)"]
        style IngestionService fill:#f0f4f8,stroke:#37474f,stroke-width:2px
        RPCServer[gRPC Server]:::rpcStyle
        HTTPListener[HTTP/Webhook Listener]:::rpcStyle
        Normalizer(Log Normalizer & Parser)
        KeywordEngine{Keyword/Regex Filter:<br/>'Error', 'Panic', 'Exception'}
        StatsEngine[Stats Aggregator]
    end

    subgraph KafkaCluster ["Kafka Event Bus"]
        style KafkaCluster fill:#fff3e0,stroke:#e65100,stroke-width:2px,stroke-dasharray: 5 5
        TopicPending[Topic: logs.to.analyze]:::kafkaStyle
        TopicResults[Topic: ai.insights.processed]:::kafkaStyle
    end

    subgraph AIWorker ["AI Processor Service (Go/Python)"]
        style AIWorker fill:#e1bee7,stroke:#4a148c
        BatchConsumer[Kafka Consumer<br/>Batch Strategy]
        OpenAIClient[OpenAI Client]
        ResultProducer[Result Publisher]
    end

    subgraph StorageLayer ["Persistence"]
        DuckDB[(DuckDB: Hot Stats)]:::dbStyle
        UserDB[(Postgres/Mongo: User Insights)]:::dbStyle
    end
    
    subgraph Presentation ["Interfaces"]
        CLI[CLI / TUI]
        Dashboard[Web Dashboard / Grafana]
    end
    
    subgraph ExtAI ["External AI"]
        OpenAI[OpenAI API]:::aiStyle
    end

    SDK -->|gRPC| RPCServer
    Cloud -->|JSON Payload| HTTPListener
    LogFile -->|Tail| Normalizer
    RPCServer --> Normalizer
    HTTPListener --> Normalizer

    Normalizer --> KeywordEngine
    KeywordEngine -->|All Logs| StatsEngine
    StatsEngine -->|Write Metrics| DuckDB
    KeywordEngine -->|If 'Error' or 'Critical' found| TopicPending

    TopicPending -->|Consume Batch| BatchConsumer
    BatchConsumer -->|Send Context| OpenAIClient
    OpenAIClient -->|Request Analysis| OpenAI
    OpenAI -->|Return RCA| OpenAIClient
    OpenAIClient -->|Format JSON| ResultProducer
    ResultProducer -->|Push Analysis| TopicResults
    TopicResults -->|Consumer| UserDB

    DuckDB -.->|Read Stats| CLI
    UserDB -.->|Read Insights| Dashboard
    UserDB -.->|Read Insights| CLI

    classDef rpcStyle fill:#d1c4e9,stroke:#512da8,stroke-width:2px;
    classDef kafkaStyle fill:#ffcc80,stroke:#ef6c00,stroke-width:2px,stroke-dasharray: 5 5;
    classDef dbStyle fill:#ffe0b2,stroke:#e65100,stroke-width:2px;
    classDef aiStyle fill:#e1bee7,stroke:#4a148c,stroke-width:2px;
```

# 🔍 Log Insight (AI-Powered Log Analysis Platform)

A distributed, event-driven log analysis platform that ingests logs from various sources (Files, Datadog, HTTP), processes them through an AI Worker (LLM) to identify root causes, and displays real-time insights in a TUI dashboard.

## 🏗 Architecture So Far
The system acts as a pipeline with three distinct components:

1.  **Ingestor Service (Go)**
    * Listens for logs via HTTP (Datadog Webhooks) or internal generators.
    * Normalizes data into a standard `LogEntry` format.
    * Pushes critical/error logs to a **Kafka** topic (`logs.to.analyze`).
    * *Status:* ✅ Working (Handles Datadog Webhooks & Manual Curls).
    - [ ] **gRPC Server:** Add a gRPC handler to the Ingestor for high-performance internal logging.
    - [ ] **Database:** Migrate from `insights.jsonl` to SQLite or DuckDB.

2.  **AI Worker (Go)**
    * Consumes messages from Kafka.
    * (Currently) Mocks AI analysis or connects to OpenAI.
    * Saves analyzed insights to a local storage file (`insights.jsonl`).
    * *Status:* ✅ Working (Consumes & Saves).

3.  **CLI Dashboard (Bubbletea TUI)**
    * Reads from storage (`insights.jsonl`).
    * Displays a navigable, interactive list of error logs and their AI-generated root causes.
    * *Status:* ✅ Working (Basic view).
    * Added polling to the CLI so new alerts appear without restarting the app.

---

## 🎯 End Goal
To build an "Agentic" Observability Platform that doesn't just show logs, but **understands** them.
* **True Agentic Behavior:** RPC agents sitting on user servers pushing logs directly.
* **Real Intelligence:** OpenAI/LLM analyzing complex stack traces.
* **Custom workflows and recreating the issues:** here once the domain system receveis an error devs can setup custom workflows that will run to recover the issue , recreting and reassuring the issue and other triggers and stuff people worked on that part of the system agets triggered sand stuff like maybe n8n flows.
* **Production Storage:** Replacing flat files with high-performance DBs (DuckDB/ClickHouse).

---

## ✅ To-Do List (Next Steps)

### Phase 1: Robustness (Immediate)
- [ ] **Real AI Connection:** Switch Worker from "Mock Analysis" to real OpenAI API calls.

### Phase 2: Agentic Expansion
- [ ] **SDK:** Build a tiny Go SDK that developers can import to send logs to us automatically.

### Phase 3: Infrastructure
- [ ] **Dockerize:** Create a full `docker-compose` for the Ingestor, Worker, and DB.

---

## 🚀 Quick Start
1. **Start Infrastructure:** `docker-compose up` (Kafka/Zookeeper)
2. **Start Ingestor:** `go run cmd/ingestor/main.go`
3. **Start Worker:** `go run cmd/worker/main.go`
4. **Start Tunnel:** `ngrok http 8080` (for Datadog)
5. **View Dashboard:** `go run cmd/cli/main.go`

Agentic Debugger

```mermaid
graph TD
    subgraph UserSpace ["User Environment"]
        %% FIXED: Quotes around Node Label
        Codebase["User Codebase<br/>(Local Dir or GitHub)"]
        App[User Application]
    end

    subgraph DataPipeline ["Log Pipeline"]
        Ingestor[Ingestor Service]
        Kafka[Kafka: logs.to.analyze]
    end

    subgraph AgenticCore ["Agentic AI Core"]
        Indexer[Code Indexer CLI]
        VectorDB[(Qdrant: Vector Store)]
        Worker[AI Worker]
        %% FIXED: Quotes around Node Label
        OpenAI["OpenAI API<br/>(Embeddings & Chat)"]
    end

    subgraph Storage ["Persistence"]
        SQLite[(SQLite: Insights)]
    end

    %% Flow
    Codebase -->|1. Scan & Chunk| Indexer
    Indexer -->|2. Generate Embeddings| OpenAI
    OpenAI -->|3. Vectors| Indexer
    Indexer -->|4. Store Vectors| VectorDB

    App -->|5. Push Error Log| Ingestor
    Ingestor --> Kafka
    Kafka -->|6. Consume Log| Worker
    
    %% FIXED: Added quotes inside the pipes |"..."|
    Worker -->|"7. Similarity Search (Error Msg)"| VectorDB
    VectorDB -->|8. Return Relevant Code Snippets| Worker
    
    %% FIXED: Added quotes inside the pipes |"..."|
    Worker -->|"9. Send (Log + Code Snippets)"| OpenAI
    OpenAI -->|10. Return Fix Suggestion| Worker
    
    Worker -->|11. Save Insight| SQLite
```



Flow for creating embeddings [High]

```mermaid
flowchart TD
    subgraph Client_Side [Client Side / Local Machine]
        direction TB
        Step1[<b>Step 1: Code Chunking</b><br/>Split code into semantic chunks]
        Step2[<b>Step 2: Merkle Tree Construction</b><br/>Compute hash tree of all valid files]
        Step5[<b>Step 5: Periodic Updates</b><br/>Check for changes every 10 mins<br/>using Merkle tree]
        
        Step1 --> Step2
        Step2 -.-> Step5
    end

    subgraph Server_Side [Server Side Processing]
        direction TB
        Step3[<b>Step 3: Embedding Generation</b><br/>Create vector representations]
        Step4[<b>Step 4: Storage and Indexing</b><br/>Store in vector database]
        DB[(<b>Turbopuffer</b><br/>Vector Database)]
        
        Step3 --> Step4
        Step4 --> DB
    end

    subgraph Usage [RAG Usage]
        UserQuery([User: @Codebase or CMD+K])
        RAG[<b>RAG Retrieval</b><br/>Retrieve relevant chunks for context]
    end

    %% Connections across subgraphs
    Step2 -- "Sync (Initial)" --> Step3
    Step5 -- "Sync (Only changed files)" --> Step3
    
    %% Usage flow
    UserQuery --> RAG
    DB -.-> RAG
    
    %% Styling
    classDef blue fill:#e1f5fe,stroke:#01579b,stroke-width:2px;
    classDef green fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;
    classDef orange fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef purple fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px;
    classDef db fill:#0277bd,stroke:#01579b,stroke-width:2px,color:white;
    
    class Step1,Step4 blue;
    class Step2 green;
    class Step3 orange;
    class Step5 purple;
    class DB db;
```