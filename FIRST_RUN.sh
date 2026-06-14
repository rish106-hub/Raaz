#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════╗"
echo "║  Raaz - First Run Setup                                ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

echo "📦 Verifying Go installation..."
go version
echo ""

echo "📦 Verifying Docker installation..."
docker --version
echo ""

echo "✅ Your system is ready!"
echo ""
echo "═════════════════════════════════════════════════════════"
echo ""
echo "🚀 TO GET STARTED:"
echo ""
echo "Option 1: Run locally (no database needed)"
echo "  $ make run"
echo ""
echo "Option 2: Run with full stack (PostgreSQL + Redis)"
echo "  $ make docker-up"
echo ""
echo "📚 Read these docs:"
echo "  • QUICKSTART.txt        ← Start here!"
echo "  • PROJECT_OVERVIEW.md   ← Architecture & features"
echo "  • SETUP.md              ← Comprehensive guide"
echo ""
echo "🧪 Verify everything works:"
echo "  $ make test"
echo ""
echo "═════════════════════════════════════════════════════════"
