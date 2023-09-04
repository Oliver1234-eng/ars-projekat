# A centralized service configuration system has been implemented. The system consists of two main components: a web service that accepts user requests and performs processing, and a database that stores the system's state. It also has two auxiliary components that maintain the system: components for storing and viewing logs and traces, and components for storing and viewing metrics.
# The web service is implemented using the Go programming language (Golang). The service provides the following operations: </br> - adding configuration to the system, where configuration is accepted as JSON data. </br> - adding a configuration group, where a group can have 1 or more configurations, and the configuration group is accepted as JSON data. </br> - viewing configuration, where configuration is retrieved by identifier. </br> - viewing a configuration group, where a group is retrieved by identifier. </br> - deleting configuration, where configuration is deleted by identifier. </br> - deleting configuration groups, where a group is deleted by identifier. </br> - expanding a configuration group, adding new configurations within the configuration group. </br> - advanced operations on the configuration group using a label system. 
# Each configuration within the configuration group have a set of labels used for filtering and searching. Multiple configurations within a group can have the same set of labels. Labels are textual pairs in the format key:value separated by ; (l1:v1;l2:v2, ...). When a user wants to retrieve configurations within a configuration group using labels, all labels in the query must match those associated with the configuration. Deletion using the label system is supported, and the same rules apply as for searching.
# Immutability is enabled, meaning there is no partial configuration modification - configuration can only be replaced entirely. Idempotent requests are supported. UUIDs are used as unique identifiers. Versioning is enabled, allowing configurations to be stored in different versions. When a client requests configuration, they must specify the version of the configuration they want to receive. When a client requests a configuration group, they must also specify the version of the group they want to receive.
# Configurations and configuration groups are stored in the NoSQL database Consul. Information about request idempotence is also stored in the NoSQL database Consul.
# The service and the database are containerized using Docker - multi-stage build. Tracing is supported in the service. Requests in the service are counted. All components are launched within a Docker Compose. The service can be tested using Postman or cURL.
