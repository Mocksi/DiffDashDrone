## TODO:

- [ ] Add a --help option
- [ ] Add a --version option
- [ ] Create a Storage package
- [ ] Create an Analyzer package
- [ ] Create a Reporter package
- [ ] Hook up to the LLM cli
- [ ] Make magic happen

## Queries:

```sql
SELECT count(\*) as ctn, filename FROM commits WHERE regexp_matches(message, '(fix?|fix(es|ed)?|close(s|d)?|revert(s|d)?)') GROUP BY filename ORDER BY ctn DESC;
```
