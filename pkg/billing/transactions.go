package billing

const generateTransactions = `
FOR service IN @@services // Iterate over Services
    LET account = LAST( // Find Service owner Account
    FOR node, edge, path IN 2
    INBOUND service
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )
    
    LET records = ( // Collect all unprocessed records
        FOR record IN @@records
        FILTER record.end != null
        FILTER !record.processed
        FILTER record.instance IN service.instances
            UPDATE record._key WITH { processed: true } IN @@records RETURN NEW
    )
    
    FILTER LENGTH(records) > 0 // Skip if no Records (no empty Transaction)
    INSERT {
        exec: DATE_NOW() / 1000, // Timestamp in seconds
        processed: false,
        account: account._key,
        service: service._key,
        records: records[*]._key,
        total: SUM(records[*].total) // Calculate Total
    } IN @@transactions RETURN NEW
`
