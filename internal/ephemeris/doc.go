// SPDX-License-Identifier: AGPL-3.0-only

// Package ephemeris implements Swiss Ephemeris 2.10.03 with native Go records,
// slices and file I/O. State owns all mutable calculation data. Large harmonic,
// precession and eclipse coefficient tables are shared read-only.
//
// The modules follow the original numerical domains: calendar, houses, JPL,
// lunar and planetary series, apparent positions, coordinate/time mathematics,
// eclipses and heliacal visibility. Private names retain the original function
// names in camel case, making the pinned upstream source easy to consult.
//
// Angles in the low-level mathematics are usually radians; the public Swiss
// API uses degrees unless its flags select radians or Cartesian coordinates.
// Dates are Julian days, with UT/TT made explicit at the API boundary. Numeric
// slices contain the documented public outputs or contiguous series values;
// they never encode addresses or object layouts.
//
// Preserve operation order, signed/unsigned integer conversions and rounded
// constants when editing a series. Tiny position changes can be amplified by
// finite differences used for velocities. The C and Flutter differential suites
// are the regression contract, with their existing field-specific tolerances.
package ephemeris
