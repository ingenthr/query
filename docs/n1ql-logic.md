# N1QL&mdash;Query Language for N1NF (Non-1st Normal Form): Logic

* Status: DRAFT/PROPOSAL
* Latest: [n1ql-logic](https://github.com/couchbaselabs/query/blob/master/docs/n1ql-logic.md)
* Modified: 2013-07-13

## Summary

N1QL is a query language for Couchbase, and is a continuation of
[UNQL](https://github.com/couchbaselabs/tuqqedin/blob/master/docs/unql-2013.md).
This document builds on the [N1QL Select
spec](https://github.com/couchbaselabs/query/blob/master/docs/n1ql-select.md)
and the [N1QL DML spec]
(https://github.com/couchbaselabs/query/blob/master/docs/n1ql-dml.md).

This document describes the syntax and semantics of the logic
statements in the language.  These are sometimes called procedural or
compound statements in other datqbase systems.  Stored logic
(procedures and functions) is defined in a separate spec.

## Motivation

There are several reasons for providing these logic capabilities in
the query language::

* Application developers sometimes want to push logic and processing down into the database.  This could be to reduce network traffic, to leverage database hardware, and to avoid complexities and risks of managing connections, connectivity, etc.
* We could (and might) also embed host languages in the database server to get similar logic capabilities.  However, application developers would still be left with the impedance mismatches between the query language and the host language (our host language bindings might mitigate this).  By providing logic statements in the query language, it is very natural and convenient for application developers to combine queries and logic.
* These capabilities are well validated.  All major database systems provide logic capabilities that are actively used.
* N1QL borrows some features from Go and other newer programming languages.  These features may distinguish N1QL from other procedural database languages.

## Datasets

N1QL DML statements allow for mutations of entire documents, and of
fragments within documents.  As such, N1QL DML statements use
path-based datasets to identify the data to be mutated.  The dataset
can be as simple as a bucket name, in which case entire documents will
be mutated; or as complex as a nested path within each document, in
which case the actual document fragments identified by the nested path
will be mutated within each document.

dataset:

![](diagram/dataset.png)

## UPDATE

The UPDATE statement has the additional ability to UNSET fields, which
removes the entire field, including its name, from the containing
object.

update:

![](diagram/update.png)

## DELETE

The DELETE statement can delete entire documents or fragments within
documents.  When deleting fragments, the DELETE statement does not
UNSET fields in the containing object.  Array-valued fields are left
as empty arrays, and scalar-valued fields are set to NULL.  In other
words, DELETE is always schema-safe for containing objects (unless
NULL is a schema violation).

delete:

![](diagram/delete.png)

## INSERT-VALUES

INSERT-VALUES inserts documents or fragments specified in its VALUES
clause.

If the dataset identifies a bucket, each expr in VALUES is inserted as
a new document in that bucket.

If the dataset identifies a path within documents, the expr VALUES are
evaluated in the context of each document and inserted as fragments at
that location within each matching document.

insert-values:

![](diagram/insert-values.png)

## INSERT-SELECT

INSERT-SELECT is currently limited to bucket-scope inserts of full
documents.  This is to prevent users from inadvertently running
correlated subqueries.

insert-select:

![](diagram/insert-select.png)

## About this Document

The
[grammar](https://github.com/couchbaselabs/query/blob/master/docs/n1ql-logic.ebnf)
forming the basis of this document is written in a [W3C dialect of
EBNF](http://www.w3.org/TR/REC-xml/#sec-notation).

This grammar has not yet been converted to an actual implementation,
ambiguities and conflicts may still be present.

Diagrams were generated by [Railroad Diagram
Generator](http://railroad.my28msec.com/) ![](diagram/.png)

### Document History

* 2013-07-07 - Initial checkin
    * UPDATE, DELETE, INSERT-VALUES, and INSERT-SELECT
    * Datasets
    * WHERE clauses in INSERT statements
    * RETURNING clauses
* 2013-07-08 - Added to open issues

### Open Issues

This meta-section records open issues in this document, and will
eventually disappear.

1.  DELETE of array-terminated paths deletes elements inside the
array, potentially leaving an empty containing array.  DELETE of
scalar-terminated paths sets the field to NULL.  Is this consistent?

1.  INSERT into array-terminated paths inserts into the array; INSERT
into scalar-terminated paths converts the field to an array.  Should
it be disallowed instead?  DELETE and INSERT should behave somewhat
inversely, and certainly predictably and intuitively.

1.  Should we provide UPDATE-APPEND?  This would be similar to
INSERT-VALUES over array-valued datasets.  INSERT seems to be closer
to the spirit of N1NF, i.e. embedded datasets are first-class
entities.

1.  Should we provide UPDATE SET VALUE() = expr for mutating entire
documents?  It would violate the nice property that function calls are
always read-only.

1.  Should we require that datasets terminate in an array-valued
field?  This would clarify / simplify most of the issues above.  Ditto
for SELECT.