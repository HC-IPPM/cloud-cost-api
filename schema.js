
import { addMocksToSchema } from '@graphql-tools/mock'
import { makeExecutableSchema } from '@graphql-tools/schema'
import casual from 'casual'

const typeDefinitions = /* GraphQL */ `
scalar Name
scalar Acro
scalar Float

type Cost {
    lastSixMonths: [Float] # For visualization purposes
    currentMonthDeltaPercentage: Float # -32.00
    previousMonth: Float
    previousFiscalYear: Float
    # TODO: Ugh so awkward. Find a better name for this.
    currentFiscalYearDeltaPercentage: Float # -32.00 
    previousCalendarYear: Float
    currentMonthToDate: Float
    currentFiscalToDate: Float
    currentFiscalEstimated: Float
    currentCalendarYearToDate: Float
    currentCalendarYearEstimated: Float
}

enum CLASSIFICATION {
    PROTECTED_B
    PROTECTED_A
    UNCLASSIFIED
}

type OrganizationalUnit {
    name: Name!
    acronym: Acro!
    costs: Cost
    projects: [Project]
}

type Project {
    id: ID!
    costs: Cost
    classification: CLASSIFICATION
    owner: OrganizationalUnit    
}


type Query {
  allProjects: [Project]
  project(id: String): Project
  projects(id: [String]): [Project]
}`

const mocks = {
    String: () => 'Mock Data',
    Float: () => {
        return (casual.double(0,500)).toFixed(2)
    },
    Name: () => {
        return casual.random_element(['Product Innovation', 'Infrastructure Management', 'Product Management', 'Infrastructure Sustainment', 'Cloud Services and Infrastructure'])
    },
    Acro: () => {
        const acro = casual._letter() + casual._letter() + casual._letter()
        return acro.toUpperCase()
    }
}
export const schema = addMocksToSchema({
    schema: makeExecutableSchema({
        typeDefs: [typeDefinitions]
    }),
    mocks: mocks
})