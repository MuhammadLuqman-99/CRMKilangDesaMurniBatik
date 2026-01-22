// CRM Kilang Desa Murni Batik - MongoDB Initialization Script
// ===========================================================

// Switch to admin database for authentication
db = db.getSiblingDB('admin');

// Create application user for crm_customers database
db.createUser({
    user: 'crm_app',
    pwd: 'crm_app_password',
    roles: [
        {
            role: 'readWrite',
            db: 'crm_customers'
        }
    ]
});

// Switch to customers database
db = db.getSiblingDB('crm_customers');

// Create collections with validation schemas
db.createCollection('customers', {
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['tenant_id', 'name', 'type', 'created_at', 'updated_at'],
            properties: {
                tenant_id: {
                    bsonType: 'string',
                    description: 'Tenant ID is required'
                },
                name: {
                    bsonType: 'string',
                    minLength: 1,
                    maxLength: 255,
                    description: 'Customer name is required'
                },
                type: {
                    enum: ['individual', 'business'],
                    description: 'Customer type must be individual or business'
                },
                email: {
                    bsonType: 'string',
                    description: 'Customer email'
                },
                phone: {
                    bsonType: 'string',
                    description: 'Customer phone'
                },
                addresses: {
                    bsonType: 'array',
                    items: {
                        bsonType: 'object',
                        properties: {
                            type: {
                                enum: ['billing', 'shipping', 'home', 'office']
                            },
                            street: { bsonType: 'string' },
                            city: { bsonType: 'string' },
                            state: { bsonType: 'string' },
                            postal_code: { bsonType: 'string' },
                            country: { bsonType: 'string' },
                            is_primary: { bsonType: 'bool' }
                        }
                    }
                },
                tags: {
                    bsonType: 'array',
                    items: { bsonType: 'string' }
                },
                custom_fields: {
                    bsonType: 'object'
                },
                status: {
                    enum: ['active', 'inactive', 'archived'],
                    description: 'Customer status'
                },
                created_at: {
                    bsonType: 'date',
                    description: 'Creation timestamp is required'
                },
                updated_at: {
                    bsonType: 'date',
                    description: 'Update timestamp is required'
                },
                deleted_at: {
                    bsonType: ['date', 'null'],
                    description: 'Soft delete timestamp'
                }
            }
        }
    }
});

// Create contacts collection
db.createCollection('contacts', {
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['tenant_id', 'customer_id', 'first_name', 'last_name', 'created_at', 'updated_at'],
            properties: {
                tenant_id: {
                    bsonType: 'string',
                    description: 'Tenant ID is required'
                },
                customer_id: {
                    bsonType: 'string',
                    description: 'Customer ID is required'
                },
                first_name: {
                    bsonType: 'string',
                    minLength: 1,
                    maxLength: 100,
                    description: 'First name is required'
                },
                last_name: {
                    bsonType: 'string',
                    minLength: 1,
                    maxLength: 100,
                    description: 'Last name is required'
                },
                email: {
                    bsonType: 'string',
                    description: 'Contact email'
                },
                phone_numbers: {
                    bsonType: 'array',
                    items: {
                        bsonType: 'object',
                        properties: {
                            type: { enum: ['mobile', 'work', 'home', 'fax'] },
                            number: { bsonType: 'string' },
                            is_primary: { bsonType: 'bool' }
                        }
                    }
                },
                position: {
                    bsonType: 'string',
                    description: 'Job position/title'
                },
                department: {
                    bsonType: 'string',
                    description: 'Department'
                },
                social_profiles: {
                    bsonType: 'object',
                    properties: {
                        linkedin: { bsonType: 'string' },
                        twitter: { bsonType: 'string' },
                        facebook: { bsonType: 'string' },
                        instagram: { bsonType: 'string' }
                    }
                },
                is_primary: {
                    bsonType: 'bool',
                    description: 'Is primary contact'
                },
                notes: {
                    bsonType: 'string'
                },
                created_at: {
                    bsonType: 'date'
                },
                updated_at: {
                    bsonType: 'date'
                },
                deleted_at: {
                    bsonType: ['date', 'null']
                }
            }
        }
    }
});

// Create indexes for customers collection
db.customers.createIndex({ 'tenant_id': 1 });
db.customers.createIndex({ 'tenant_id': 1, 'email': 1 }, { unique: true, sparse: true });
db.customers.createIndex({ 'tenant_id': 1, 'name': 'text' });
db.customers.createIndex({ 'tenant_id': 1, 'type': 1 });
db.customers.createIndex({ 'tenant_id': 1, 'status': 1 });
db.customers.createIndex({ 'tenant_id': 1, 'tags': 1 });
db.customers.createIndex({ 'created_at': 1 });
db.customers.createIndex({ 'updated_at': 1 });

// Create indexes for contacts collection
db.contacts.createIndex({ 'tenant_id': 1 });
db.contacts.createIndex({ 'tenant_id': 1, 'customer_id': 1 });
db.contacts.createIndex({ 'tenant_id': 1, 'email': 1 }, { unique: true, sparse: true });
db.contacts.createIndex({ 'tenant_id': 1, 'first_name': 'text', 'last_name': 'text' });
db.contacts.createIndex({ 'created_at': 1 });

// Create activity_logs collection for audit trail
db.createCollection('activity_logs', {
    capped: true,
    size: 104857600, // 100MB
    max: 1000000
});

db.activity_logs.createIndex({ 'tenant_id': 1 });
db.activity_logs.createIndex({ 'entity_type': 1, 'entity_id': 1 });
db.activity_logs.createIndex({ 'user_id': 1 });
db.activity_logs.createIndex({ 'created_at': 1 }, { expireAfterSeconds: 7776000 }); // 90 days TTL

print('MongoDB initialization completed successfully!');
