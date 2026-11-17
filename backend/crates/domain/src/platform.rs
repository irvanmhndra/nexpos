//! Who on the internal team may do what in the platform panel.

/// Failed sign-ins in a row before an account locks.
pub const MAX_FAILED_LOGINS: i32 = 5;
/// How long a locked account stays locked.
pub const LOCK_MINUTES: i32 = 15;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Role {
    /// Everything, including other staff accounts. Always sees every company.
    Admin,
    /// Runs companies: may suspend and reactivate them.
    Ops,
    /// Helps customers: looks, and opens read-only sessions.
    Support,
}

impl Role {
    pub fn parse(text: &str) -> Option<Self> {
        match text {
            "admin" => Some(Self::Admin),
            "ops" => Some(Self::Ops),
            "support" => Some(Self::Support),
            _ => None,
        }
    }

    pub fn as_str(self) -> &'static str {
        match self {
            Self::Admin => "admin",
            Self::Ops => "ops",
            Self::Support => "support",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Permission {
    ViewCompanies,
    SuspendCompany,
    Impersonate,
    ViewAudit,
    ManageStaff,
}

pub fn allows(role: Role, permission: Permission) -> bool {
    use Permission::*;
    match role {
        Role::Admin => true,
        Role::Ops => matches!(
            permission,
            ViewCompanies | SuspendCompany | Impersonate | ViewAudit
        ),
        Role::Support => matches!(permission, ViewCompanies | Impersonate | ViewAudit),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn roles_round_trip() {
        for role in [Role::Admin, Role::Ops, Role::Support] {
            assert_eq!(Role::parse(role.as_str()), Some(role));
        }
        assert_eq!(Role::parse("owner"), None);
    }

    #[test]
    fn only_admin_manages_staff_and_support_cannot_suspend() {
        assert!(allows(Role::Admin, Permission::ManageStaff));
        assert!(!allows(Role::Ops, Permission::ManageStaff));
        assert!(!allows(Role::Support, Permission::ManageStaff));
        assert!(allows(Role::Ops, Permission::SuspendCompany));
        assert!(!allows(Role::Support, Permission::SuspendCompany));
        assert!(allows(Role::Support, Permission::Impersonate));
    }
}
