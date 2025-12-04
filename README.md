# otter: graph-based authorization

## Basic Concepts

```mermaid
graph BT;

subgraph Resources
    r0@{ shape: dbl-circ, label: "_" }
    r1@{ shape: circle, label: "R1" }
    r2@{ shape: circle, label: "R2" }
    r3@{ shape: circle, label: "R3" }

    r1 -->|CHILD_OF| r0
    r2 -->|CHILD_OF| r0
    r3 -->|CHILD_OF| r2
end


subgraph Subjects
    g1@{ shape: circle, label: "Group1" }
    g2@{ shape: circle, label: "Group2" }
    p1@{ shape: circle, label: "Principal1" }
    p2@{ shape: circle, label: "Principal2" }

    p1 -->|CHILD_OF| g1
    p2 -->|CHILD_OF| g2
    g1 -->|CHILD_OF| g2
end

subgraph Specifiers
    s0@{ shape: dbl-circ, label: "* = *" }

    sEnvRoot@{ shape: dbl-circ, label: "env = *" }
    sEnvProd@{ shape: circle, label: "env = prod" }
    sEnvDev@{ shape: circle, label: "env = dev" }

    sRoleRoot@{ shape: dbl-circ, label: "role = *" }
    sRoleAdmin@{ shape: circle, label: "role = admin" }
    sRoleUser@{ shape: circle, label: "role = user" }

    sEnvRoot -->|CHILD_OF| s0
    sEnvDev -->|CHILD_OF| sEnvRoot
    sEnvProd -->|CHILD_OF| sEnvRoot

    sRoleRoot -->|CHILD_OF| s0
    sRoleAdmin -->|CHILD_OF| sRoleRoot
    sRoleUser -->|CHILD_OF| sRoleAdmin
end

pc1@{ shape: rounded, label: "Policy1" }

r2 ==>|HAS_POLICY| pc1
g1 ==>|HAS_POLICY| pc1
pc1 ==>|READ| sEnvDev
pc1 ==>|READ| sRoleRoot
```

### Subject
Representation of a user/group of users. Represents the `Who`.\
`(:Subject {name: "<unique-id>", type: "<Principal | Group>"})`

Connected to other subjects using a hierarchy.\
`(child)-[:CHILD_OF]->(parent)`

**Assumptions**:
- Principals can only be children of Groups, not other Principals.

### Resource
Object that needs to be authorized. Represents the `What`\
`(:Resource {name: "<unique-id>"})`

Connected to other resources using a hierarchy.\
`(child)-[:CHILD_OF]->(parent)`

**Assumptions**:
- Root resource has `name: "_"`. Does not have any parents.

### Specifier
Defines additional properties for the permission.\
More extensible version of adding the properties to the graph edge.\
`(:Specifier {key: "<key>", value: "<value"})`

Connected to other specifiers using a hierarchy.\
`(child)-[:CHILD_OF]->(parent)`


**Assumptions**:
- Root specifier has `key: "*", value: "*"`.
- Immediate children have `key: "<key>", value: "*"`.

### Action
`How` a particular `Subject` can access a `Resource`.\
Represented in the graph as the edge type between a `Policy` and `Specifier` node.

### Policy
Intermediate node to represent a permission. Needed since the subject and resource are to be used as one group.\
`(:Policy {id: "<uuid>"})`

Policies are represented as:
```
(policy:Policy {id: "<uuid>"})
(subject:Subject)-[:HAS_POLICY]->(policy)
(resource:Resource)-[:HAS_POLICY]->(policy)
(policy)-[:<action>]->(specifier:Specifier)
```

## Querying
### Can
`Can <Subject> perform <Action> on <Resource> with <Specifiers>?`\
Yes/No question, returns a boolean.

Give all details, check if there is a path.

### WhatCan
`WhatCan <Subject> perform <Action> on with <Specifiers> [under <Parent Resource>]?`\
List of resources, bounded by the optional parent resource in the hierarchy.

Fetch resources given everything else.

### WhoCan
`WhoCan <Action> on <Resource> with <Specifiers>?`\
List of subjects.

Fetch subjects given everything else.

### HowCan
`HowCan <Subject> perform <Action> on <Resource> [with <Specifiers>]?`\
Fetch specifiers given everything else. Optionally provide specifiers to reduce output space.

This is a heavy query since it returns a cartesian product of all applicable specifiers.