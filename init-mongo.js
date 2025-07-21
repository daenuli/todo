// MongoDB initialization script
// This script runs when the MongoDB container starts for the first time

// Switch to the todoapp database
db = db.getSiblingDB('todoapp');

// Create a user for the todoapp database
db.createUser({
  user: 'todouser',
  pwd: 'todopass',
  roles: [
    {
      role: 'readWrite',
      db: 'todoapp'
    }
  ]
});

// Create the todos collection with some sample data (optional)
db.todos.insertMany([
  {
    title: "Welcome to Todo API",
    description: "This is a sample todo item created during database initialization",
    completed: false,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

print('Database initialization completed!');