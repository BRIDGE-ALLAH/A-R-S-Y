// Package relations validates model-to-model relation metadata without
// executing joins, relation loading, or database schema changes.
// Validation uses model metadata only and leaves persistence to later layers.
// It also derives safe, repeatable many-to-many join-table ownership.
package relations
