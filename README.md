# Log Insight (Draft)
# 

![Architecture](content/arch.png)

![Datadog](content/dd.png)

![Ingestor Log](content/ingestor_log.png)

![Worker Log](content/worker_log.png)


## System Architecture (Mermaid)

```mermaid
graph TD
    %% =========================================
    %% 1. SOURCES (Expanded)
    %% =========================================
    subgraph Sources ["External Sources"]
        LogFile[File Tail / Syslog]
        SDK[User Apps SDK]
        Cloud[CloudWatch / Datadog Webhooks]
    end

    %% =========================================
    %% 2. INGESTION SERVICE (Go Application)
    %% =========================================
    subgraph IngestionService ["Ingestion Service (Go)"]
        style IngestionService fill:#f0f4f8,stroke:#37474f,stroke-width:2px

        %% Interfaces
        RPCServer[gRPC Server]:::rpcStyle
        HTTPListener[HTTP/Webhook Listener]:::rpcStyle
        
        %% Core Logic
        Normalizer(Log Normalizer & Parser)
        KeywordEngine{Keyword/Regex Filter:<br/>'Error', 'Panic', 'Exception'}
        
        %% Hot Path (Fast)
        StatsEngine[Stats Aggregator]
    end

    %% =========================================
    %% 3. MESSAGE BROKER (Kafka)
    %% =========================================
    subgraph KafkaCluster ["Kafka Event Bus"]
        style KafkaCluster fill:#fff3e0,stroke:#e65100,stroke-width:2px,stroke-dasharray: 5 5
        
        TopicPending[Topic: logs.to.analyze]:::kafkaStyle
        TopicResults[Topic: ai.insights.processed]:::kafkaStyle
    end

    %% =========================================
    %% 4. AI WORKER SERVICE (Decoupled)
    %% =========================================
    subgraph AIWorker ["AI Processor Service (Go/Python)"]
        style AIWorker fill:#e1bee7,stroke:#4a148c
        
        BatchConsumer[Kafka Consumer<br/>Batch Strategy]
        OpenAIClient[OpenAI Client]
        ResultProducer[Result Publisher]
    end

    %% =========================================
    %% 5. STORAGE & PRESENTATION
    %% =========================================
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

    %% =========================================
    %% CONNECTIONS
    %% =========================================

    %% Ingestion
    SDK -->|gRPC| RPCServer
    Cloud -->|JSON Payload| HTTPListener
    LogFile -->|Tail| Normalizer
    RPCServer --> Normalizer
    HTTPListener --> Normalizer

    %% Internal Processing
    Normalizer --> KeywordEngine
    
    %% Fork: Hot Path (Metrics)
    KeywordEngine -->|All Logs| StatsEngine
    StatsEngine -->|Write Metrics| DuckDB
    
    %% Fork: Cold Path (AI)
    KeywordEngine -->|If 'Error' or 'Critical' found| TopicPending
    
    %% AI Workflow
    TopicPending -->|Consume Batch| BatchConsumer
    BatchConsumer -->|Send Context| OpenAIClient
    OpenAIClient -->|Request Analysis| OpenAI
    OpenAI -->|Return RCA| OpenAIClient
    OpenAIClient -->|Format JSON| ResultProducer
    ResultProducer -->|Push Analysis| TopicResults

    %% Storage of Insights
    TopicResults -->|Consumer| UserDB

    %% Presentation
    DuckDB -.->|Read Stats| CLI
    UserDB -.->|Read Insights| Dashboard
    UserDB -.->|Read Insights| CLI

    %% =========================================
    %% STYLES
    %% =========================================
    classDef rpcStyle fill:#d1c4e9,stroke:#512da8,stroke-width:2px;
    classDef kafkaStyle fill:#ffcc80,stroke:#ef6c00,stroke-width:2px,stroke-dasharray: 5 5;
    classDef dbStyle fill:#ffe0b2,stroke:#e65100,stroke-width:2px;
    classDef aiStyle fill:#e1bee7,stroke:#4a148c,stroke-width:2px;

