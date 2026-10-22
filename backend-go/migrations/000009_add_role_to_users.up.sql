-- Add role_id to users table
ALTER TABLE users ADD COLUMN role_id BIGINT REFERENCES roles(id);

CREATE INDEX idx_users_role_id ON users(role_id);

-- Set existing users to 'owner' role (first user of each company is owner)
UPDATE users u
SET role_id = (SELECT id FROM roles WHERE code = 'owner' AND company_id IS NULL LIMIT 1);
