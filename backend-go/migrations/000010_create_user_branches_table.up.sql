-- User-Branch assignment for branch-level access control
CREATE TABLE IF NOT EXISTS user_branches (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_branches_unique ON user_branches(user_id, branch_id);
CREATE INDEX idx_user_branches_user_id ON user_branches(user_id);
CREATE INDEX idx_user_branches_branch_id ON user_branches(branch_id);

-- Ensure only one default branch per user
CREATE UNIQUE INDEX idx_user_branches_default ON user_branches(user_id) WHERE is_default = true;

-- Assign existing users to their company's default branch (HQ)
INSERT INTO user_branches (user_id, branch_id, is_default)
SELECT u.id, b.id, true
FROM users u
JOIN branches b ON b.company_id = u.company_id AND b.code = 'HQ';
