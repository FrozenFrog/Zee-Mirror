import { Fragment, ReactNode } from 'react'

// Fills {name} placeholders of an already translated string with React
// nodes, so styled fragments (bold counts, links) keep the word order of each
// language. Unknown placeholders are left as text.
export const rich = (template: string, nodes: Record<string, ReactNode>): ReactNode[] =>
  template
    .split(/(\{\w+\})/)
    .filter((part) => part !== '')
    .map((part, i) => {
      const name = /^\{(\w+)\}$/.exec(part)?.[1]
      return <Fragment key={i}>{name && name in nodes ? nodes[name] : part}</Fragment>
    })
