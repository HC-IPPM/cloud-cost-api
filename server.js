import { ApolloServer } from '@apollo/server'
import { startStandaloneServer } from '@apollo/server/standalone'
import { schema } from './schema.js'

const apollo = new ApolloServer({
    schema: schema,
})

const { url } = await startStandaloneServer(apollo, { listen: { port: 8080 } })
console.log(`Server is running on ${url}`)