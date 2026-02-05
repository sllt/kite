#!/bin/bash
# Release script for github.com/sllt/kite
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh v0.1.0

set -e

VERSION=${1:-v0.1.0}

echo "🚀 Releasing $VERSION for github.com/sllt/kite"
echo ""

# Step 1: Commit changes (if any)
if [[ -n $(git status --porcelain) ]]; then
    echo "📝 Committing changes..."
    git add .
    git commit -m "Rename module to github.com/sllt/kite

- Changed all import paths from gofr.dev to github.com/sllt/kite
- Updated go.mod, go.work, and documentation
- Prepared for initial release $VERSION"
fi

# Step 2: Create main module tag
echo "🏷️  Creating main module tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

# Step 3: Create submodule tags
SUBMODULES=(
    "pkg/kite/datasource/arangodb"
    "pkg/kite/datasource/cassandra"
    "pkg/kite/datasource/clickhouse"
    "pkg/kite/datasource/couchbase"
    "pkg/kite/datasource/dbresolver"
    "pkg/kite/datasource/dgraph"
    "pkg/kite/datasource/elasticsearch"
    "pkg/kite/datasource/file/azure"
    "pkg/kite/datasource/file/ftp"
    "pkg/kite/datasource/file/gcs"
    "pkg/kite/datasource/file/s3"
    "pkg/kite/datasource/file/sftp"
    "pkg/kite/datasource/influxdb"
    "pkg/kite/datasource/kv-store/badger"
    "pkg/kite/datasource/kv-store/dynamodb"
    "pkg/kite/datasource/kv-store/nats"
    "pkg/kite/datasource/mongo"
    "pkg/kite/datasource/opentsdb"
    "pkg/kite/datasource/oracle"
    "pkg/kite/datasource/pubsub/eventhub"
    "pkg/kite/datasource/pubsub/nats"
    "pkg/kite/datasource/pubsub/sqs"
    "pkg/kite/datasource/scylladb"
    "pkg/kite/datasource/solr"
    "pkg/kite/datasource/surrealdb"
)

echo "🏷️  Creating submodule tags..."
for module in "${SUBMODULES[@]}"; do
    tag="${module}/${VERSION}"
    echo "   - $tag"
    git tag -a "$tag" -m "Release ${module} $VERSION"
done

echo ""
echo "✅ All tags created!"
echo ""
echo "📋 Tags created:"
git tag -l "*${VERSION}*" | head -30
echo ""

# Step 4: Instructions for pushing
echo "🔄 To push to GitHub, run:"
echo ""
echo "   # Add remote (if not exists)"
echo "   git remote add origin git@github.com:sllt/kite.git"
echo ""
echo "   # Push code and all tags"
echo "   git push -u origin HEAD"
echo "   git push origin --tags"
echo ""
echo "   # Or push everything at once"
echo "   git push -u origin HEAD --tags"
