

### Como usar

Para iniciar o projeto siga os passos:

1. Na raiz do projeto execute o comando `docker compose up -d` e aguarde os `containers` do `mysql` e `rabbitmq` e `o app order system` iniciarem;
2. Para acessar o `rabbitmq` acesse: `http://localhost:15672` com `guest` para `username / password`;
3. Adicione a fila clicando em `Queues and Streams` no campo `name` coloque `orders` e clique em `Add queue`;
4. Criada a fila acesse ela e em `Bindings` no campo `From exchange` coloque `amq.direct`;

Com isso o sistema já está pronto para o uso, para testar existe algumas formas:

1. `REST API`:
    - No diretório `api` tem um arquivo `create_order.http` que usa um `plugin` do `vscode` de `client rest` https://github.com/Huachao/vscode-restclient basta executar as `requests` por esse arquivo.
2. `GRPC`:
    - Para utilizar o `GRPC` é necessário um `client`, o que foi utilizado nesse projeto é esse: https://github.com/ktr0731/evans mas pode consumir com algum outro caso queira, para utilizar o `Evans` execute o comando `evans -r repl` digite `call` e ao dar um espaço já irá aparecer os serviços diponíveis, `CreateOrder` cria um novo registro e `ListOrder` lista as ordens;
3. `GraphQL`:
    - Para utilizar `GraphQL` acesse `http://localhost:8080/` e cria a `mutation` para inserir dados:
    ```graphql
        mutation createOrder {
            createOrder(input: { id: "abc", Price: 12.2, Tax: 2.0 }) {
                id
                Price
                Tax
                FinalPrice
            }
        }

        query queryOrders {
            orders {
                id
                Price
                Tax
                FinalPrice
            }
        }
    ```

---

