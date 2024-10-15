
import { addMocksToSchema } from '@graphql-tools/mock'
import { makeExecutableSchema } from '@graphql-tools/schema'
import casual from 'casual'



const typeDefinitions = /* GraphQL */ `
scalar Name1
scalar Name2
scalar Name3
scalar Acro
scalar Float


type Cost {
    lastSixMonths: [Float] # For visualization purposes
    currentMonthDeltaPercentage: Float # -32.00
    previousMonth: Float
    previousFiscalYear: Float
    # TODO: Ugh so awkward. Find a better name for this.
    # Maybe a custom Percentage scalar?
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

# custom type with description string explaining the unit. 2 kg CO2e
# https://the-guild.dev/graphql/tools/docs/scalars#custom-graphqlscalartype-instance
scalar CarbonEquivalentKilograms 

type CarbonFootPrint {
    lastSixMonths: [CarbonEquivalentKilograms] # For visualization purposes
    currentMonthDeltaPercentage: Float
    previousMonth: CarbonEquivalentKilograms  # 2 kg CO2e
    previousFiscalYear: CarbonEquivalentKilograms
    # TODO: Ugh so awkward. Find a better name for this.
    currentFiscalYearDeltaPercentage: Float # -32.00 
    previousCalendarYear: CarbonEquivalentKilograms
    currentMonthToDate: CarbonEquivalentKilograms
    currentFiscalToDate: CarbonEquivalentKilograms
    currentFiscalEstimated: CarbonEquivalentKilograms
    currentCalendarYearToDate: CarbonEquivalentKilograms
    currentCalendarYearEstimated: CarbonEquivalentKilograms
}

type Project {
    id: String!
    costs: Cost
    classification: CLASSIFICATION
    owner: Department
    carbonFootprint: CarbonFootPrint
    # TODO: Surface unattended project activity recomendations (shutting down idle projects)
    # https://cloud.google.com/recommender/docs/unattended-project-recommender#overview
    # TODO: Are there other recomendations we should surface?
    # Maybe combine all the idle resource recommendations into a list?
    # Maybe combine all the overprovisioning recommendations into a list?
    # https://cloud.google.com/recommender/docs/recommenders
    # TODO: Explore ways to expose usage (inbound HTTP requests for web?)
}

interface OrganizationalUnit {
    acronym: Acro!
    costs: Cost
    projects: [Project]
    projectCount: Int
    totalCloudSpend: Float
    carbonFootprint: CarbonEquivalentKilograms
}

type Branch implements OrganizationalUnit {
    name: Name2!
    acronym: Acro!
    costs: Cost
    projects: [Project]
    projectCount: Int
    totalCloudSpend: Float
    carbonFootprint: CarbonEquivalentKilograms
    directorates: [Directorate]
    department: Department
}
type Directorate implements OrganizationalUnit {
    name: Name3!
    acronym: Acro!
    costs: Cost
    projects: [Project]
    projectCount: Int
    totalCloudSpend: Float
    carbonFootprint: CarbonEquivalentKilograms
    branch: Branch
    department: Department
}
type Department implements OrganizationalUnit {
    name: Name1!
    acronym: Acro!
    costs: Cost
    projects: [Project]
    projectCount: Int
    totalCloudSpend: Float
    carbonFootprint: CarbonEquivalentKilograms
    branches: [Branch]
    directorates: [Directorate]
}

type Query {
  allProjects: [Project]
  projects(id: [String]): [Project]
  project(id: String): Project
  branch(name: String): Branch
  directorate(name: String): Directorate
  department(name: String): Department
}`

const mocks = {
    String: () => 'Mock Data',
    Float: () => {
        return (casual.double(0,500)).toFixed(2)
    },
    Name1: () => {
        return casual.random_element(['Department D1', 'Department D2', 'Department D3', 'Department D4', 'Department D5'])
    },
    Name2: () => {
        return casual.random_element(['Branch B1', 'Department B2', 'Department B3', 'Department B4', 'Department B5'])    
    },
    Name3: () => {
        return casual.random_element(['Directorate A', 'Directorate B', 'Directorate C', 'Directorate D', 'Directorate E'])    
    },
    Acro: () => {
        const acro = casual._letter() + casual._letter() + casual._letter()
        return acro.toUpperCase()
    },
    CarbonEquivalentKilograms: () => {
        return (casual.double(0,500)).toFixed(2)
    }
}
export const schema = addMocksToSchema({
    schema: makeExecutableSchema({
        typeDefs: [typeDefinitions]
    }),
    mocks: mocks
})