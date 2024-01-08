### Notes
```bash
$ docker run -p "6333:6333" -p "6334:6334" --name "rag-openai-qdrant" --rm -d qdrant/qdrant:latest
$ poetry run jupyter notebook
$ poetry run jupyter server list
```

### References:
https://github.com/qdrant/examples/blob/master/rag-openai-qdrant/rag-openai-qdrant.ipynb