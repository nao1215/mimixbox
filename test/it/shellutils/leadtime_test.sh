LeadtimeHelp() { leadtime --help; }
LeadtimeNoSub() { leadtime; }
LeadtimeUnknownSub() { leadtime bogus --owner=a --repo=b; }
LeadtimeMissingToken() { LT_GITHUB_ACCESS_TOKEN= GITHUB_TOKEN= leadtime stat --owner=acme --repo=demo; }
LeadtimeMissingOwnerRepo() { LT_GITHUB_ACCESS_TOKEN=x leadtime stat; }
LeadtimeJSONMarkdownConflict() { LT_GITHUB_ACCESS_TOKEN=x leadtime stat --owner=a --repo=b --json --markdown; }
