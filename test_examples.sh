#!/bin/bash

# Test script for validating cooklang examples
# Tests different YAML frontmatter tag formats

echo "🧪 Testing Cooklang Parser Examples"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local file="$1"
    local description="$2"
    local expected_tags="$3"
    
    echo -e "\n${YELLOW}Testing:${NC} $description"
    echo "File: $file"
    
    # Parse the recipe and extract tags
    if ! output=$(go run cmd/cook/main.go "$file" 2>&1); then
        echo -e "${RED}❌ FAILED:${NC} Parser error"
        echo "$output"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Extract tags line
    tags_line=$(echo "$output" | grep "tags:" | head -1)
    
    if [[ -z "$tags_line" ]]; then
        if [[ "$expected_tags" == "NONE" ]]; then
            echo -e "${GREEN}✅ PASSED:${NC} No tags found (as expected)"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}❌ FAILED:${NC} No tags found, expected: $expected_tags"
            ((TESTS_FAILED++))
            return 1
        fi
    fi
    
    # Extract the actual tags value
    actual_tags=$(echo "$tags_line" | sed 's/.*tags: //')
    
    if [[ "$actual_tags" == "$expected_tags" ]]; then
        echo -e "${GREEN}✅ PASSED:${NC} Tags: '$actual_tags'"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED:${NC} Expected: '$expected_tags', Got: '$actual_tags'"
        ((TESTS_FAILED++))
    fi
}

# Change to project directory
cd "$(dirname "$0")"

echo -e "\n📋 Running tag format tests..."

# Test 1: Bracket array format
run_test "examples/array_tags.cook" \
         "Bracket array format: [ pasta, italian, comfort-food ]" \
         "pasta, italian, comfort-food"

# Test 2: YAML list format
run_test "examples/mixed_arrays.cook" \
         "YAML list format with dashes" \
         "spicy, asian, quick-meal, vegetarian"

# Test 3: Single tag
run_test "examples/single_tag.cook" \
         "Single tag value" \
         "breakfast"

# Test 4: Edge cases
echo -e "\n🔍 Testing edge cases..."

# Check empty arrays
echo -e "\n${YELLOW}Testing:${NC} Empty arrays and edge cases"
echo "File: examples/edge_cases.cook"

if output=$(go run cmd/cook/main.go "examples/edge_cases.cook" 2>&1); then
    # Check for empty tags
    if echo "$output" | grep -q "tags: $"; then
        echo -e "${GREEN}✅ PASSED:${NC} Empty array becomes empty string"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED:${NC} Empty array not handled correctly"
        ((TESTS_FAILED++))
    fi
    
    # Check spaced tags
    if echo "$output" | grep -q "spaced_tags: quick, easy"; then
        echo -e "${GREEN}✅ PASSED:${NC} Extra spaces removed correctly"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED:${NC} Spaced tags not cleaned correctly"
        ((TESTS_FAILED++))
    fi
    
    # Check single item array
    if echo "$output" | grep -q "single_item_array: dessert"; then
        echo -e "${GREEN}✅ PASSED:${NC} Single-item array handled correctly"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED:${NC} Single-item array not handled correctly"
        ((TESTS_FAILED++))
    fi
else
    echo -e "${RED}❌ FAILED:${NC} Parser error on edge cases"
    echo "$output"
    ((TESTS_FAILED+=3))
fi

# Test JSON output
echo -e "\n🔧 Testing JSON output..."
if json_output=$(go run cmd/cook/main.go "examples/mixed_arrays.cook" --json 2>&1); then
    if echo "$json_output" | grep -q '"tags": "spicy, asian, quick-meal, vegetarian"'; then
        echo -e "${GREEN}✅ PASSED:${NC} JSON output contains correct tags"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED:${NC} JSON output incorrect"
        ((TESTS_FAILED++))
    fi
else
    echo -e "${RED}❌ FAILED:${NC} JSON output error"
    echo "$json_output"
    ((TESTS_FAILED++))
fi

# Summary
echo -e "\n📊 Test Summary"
echo "==============="
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total tests:  $((TESTS_PASSED + TESTS_FAILED))"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed.${NC}"
    exit 1
fi
