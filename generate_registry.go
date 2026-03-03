// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

// materialClass is broad material family used for fallback tuning.
type materialClass uint8

const (
	materialClassDefault materialClass = iota
	materialClassTextile
	materialClassPolymer
	materialClassWood
	materialClassMetal
	materialClassGlass
	materialClassLeather
	materialClassTerrain
	materialClassMineral
	materialClassOrganic
)

// materialSeed defines generation seed values.
type materialSeed struct {
	specular      [4]float64
	emissive      [4]float64
	specularPower float64
	fresnelA      float64
	fresnelB      float64
	materialClass materialClass
	fresnelGlass  bool
}

// materialCatalog stores baseline generator material families.
var materialCatalog = map[BaseMaterial]materialSeed{
	BaseMaterialTextile: {
		materialClass: materialClassTextile,
		specular:      [4]float64{0.132647, 0.132647, 0.132647, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 55,
		fresnelA:      1.23,
		fresnelB:      0.36,
	},
	BaseMaterialPlastic: {
		materialClass: materialClassPolymer,
		specular:      [4]float64{0.075, 0.075, 0.075, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 100,
		fresnelA:      0.67,
		fresnelB:      0.70,
	},
	BaseMaterialRubber: {
		materialClass: materialClassPolymer,
		specular:      [4]float64{0.075, 0.075, 0.075, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 75.35,
		fresnelA:      1.435,
		fresnelB:      0.45,
	},
	BaseMaterialSteel: {
		materialClass: materialClassMetal,
		specular:      [4]float64{0.143333, 0.143333, 0.143333, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 80,
		fresnelA:      1,
		fresnelB:      0.7,
	},
	BaseMaterialRust: {
		materialClass: materialClassMetal,
		specular:      [4]float64{0.1425, 0.1425, 0.1425, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 100,
		fresnelA:      0.85,
		fresnelB:      0.32,
	},
	BaseMaterialWood: {
		materialClass: materialClassWood,
		specular:      [4]float64{0.09049, 0.09049, 0.09049, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 70,
		fresnelA:      0.99,
		fresnelB:      0.53,
	},
	BaseMaterialGlass: {
		materialClass: materialClassGlass,
		specular:      [4]float64{0.75, 0.75, 0.75, 0},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 500,
		fresnelA:      1.7,
		fresnelGlass:  true,
	},
	BaseMaterialLeather: {
		materialClass: materialClassLeather,
		specular:      [4]float64{0.08, 0.08, 0.08, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 70,
		fresnelA:      1,
		fresnelB:      0.7,
	},
	BaseMaterialEarth: {
		materialClass: materialClassTerrain,
		specular:      [4]float64{0.0575, 0.0575, 0.0575, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 75,
		fresnelA:      1.16,
		fresnelB:      0.25,
	},
	BaseMaterialPaper: {
		materialClass: materialClassTextile,
		specular:      [4]float64{0.056, 0.056, 0.056, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 80,
		fresnelA:      1,
		fresnelB:      0.45,
	},
	BaseMaterialConcrete: {
		materialClass: materialClassMineral,
		specular:      [4]float64{0.185001, 0.185001, 0.185001, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 80,
		fresnelA:      1.255,
		fresnelB:      0.35,
	},
	BaseMaterialStone: {
		materialClass: materialClassMineral,
		specular:      [4]float64{0.073824, 0.073824, 0.073824, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 50,
		fresnelA:      1.82,
		fresnelB:      0.45,
	},
	BaseMaterialSkin: {
		materialClass: materialClassOrganic,
		specular:      [4]float64{0.20, 0.20, 0.20, 1},
		emissive:      [4]float64{0, 0, 0, 1},
		specularPower: 80,
		fresnelA:      1.82,
		fresnelB:      0.71,
	},
}
