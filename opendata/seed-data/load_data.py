#!/usr/bin/env python3
"""
HCMC Smart City Data Loader
Parses CSV files and upserts them as NGSI-LD entities to Orion-LD Context Broker.

Uses Smart Data Models compliant entity types and proper NGSI-LD normalized format.
Each entity type uses its domain-specific @context for proper attribute resolution.

@version 2.0
@author CTU·SematX
@copyright (c) 2025 CTU·SematX. All rights reserved
@license MIT License
"""

import csv
import json
import os
import sys
import time
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError
from typing import Any

BROKER_URL = os.environ.get("BROKER_URL", "http://localhost:1026")
DATA_DIR = os.environ.get("DATA_DIR", "/data")

# Smart Data Models @context URLs by CSV filename
# Each entity type uses its domain-specific context
CONTEXT_MAP = {
    "AirQualityObserved": "https://smart-data-models.github.io/dataModel.Environment/context.jsonld",
    "WeatherObserved": "https://smart-data-models.github.io/dataModel.Weather/context.jsonld",
    "WeatherAlert": "https://smart-data-models.github.io/dataModel.Weather/context.jsonld",
    "TrafficFlowObserved": "https://smart-data-models.github.io/dataModel.Transportation/context.jsonld",
    "FloodSensor": "https://smart-data-models.github.io/dataModel.Environment/context.jsonld",
    "FloodZone": "https://smart-data-models.github.io/dataModel.Environment/context.jsonld",
    "EmergencyVehicle": "https://smart-data-models.github.io/dataModel.Transportation/context.jsonld",
    "MedicalFacility": "https://smart-data-models.github.io/dataModel.Building/context.jsonld",
}

# Entity type mapping: CSV filename -> NGSI-LD entity type
TYPE_MAP = {
    "AirQualityObserved": "AirQualityObserved",
    "WeatherObserved": "WeatherObserved",
    "WeatherAlert": "WeatherAlert",
    "TrafficFlowObserved": "TrafficFlowObserved",
    "FloodSensor": "FloodMonitoring",
    "FloodZone": "FloodMonitoring",
    "EmergencyVehicle": "Vehicle",
    "MedicalFacility": "Building",
}

# Attribute mapping: CSV column -> NGSI-LD attribute name (per entity type)
# Based on Smart Data Models normalized examples
ATTRIBUTE_MAP = {
    "AirQualityObserved": {
        "pm25": "pm25",
        "pm10": "pm10",
        "no2": "no2",
        "so2": "so2",
        "co": "co",
        "o3": "o3",
        "aqi": "airQualityIndex",
        "aqiCategory": "airQualityLevel",
    },
    "WeatherObserved": {
        "temperature": "temperature",
        "humidity": "relativeHumidity",
        "windSpeed": "windSpeed",
        "windDirection": "windDirection",
        "atmosphericPressure": "atmosphericPressure",
        "precipitation": "precipitation",
    },
    "WeatherAlert": {
        "incidentType": "subCategory",
        "severity": "severity",
        "status": "category",
    },
    "TrafficFlowObserved": {
        "averageVehicleSpeed": "averageVehicleSpeed",
        "vehicleCount": "intensity",
        "congestionIndex": "occupancy",
        "roadName": "laneId",
    },
    "FloodSensor": {
        "waterLevel": "currentLevel",
        "batteryLevel": "measuredDistance",
    },
    "FloodZone": {
        "waterDepth": "currentLevel",
        "floodSeverity": "floodLevelStatus",
        "affectedPopulation": "stationID",
        "areaType": "alertLevel",
        "isActive": "dangerLevel",
    },
    "EmergencyVehicle": {
        "vehicleType": "vehicleType",
        "speed": "speed",
        "heading": "bearing",
        "status": "serviceStatus",
    },
    "MedicalFacility": {
        "bedCapacity": "floorsAboveGround",
        "availableBeds": "floorsBelowGround",
    },
}

# Retry configuration
MAX_RETRIES = 30
RETRY_DELAY = 2


def wait_for_broker():
    """Wait for broker to be ready."""
    print(f"Waiting for broker at {BROKER_URL}...")
    for i in range(MAX_RETRIES):
        try:
            req = Request(f"{BROKER_URL}/version")
            with urlopen(req, timeout=5) as resp:
                if resp.status == 200:
                    print("Broker is ready!")
                    return True
        except (HTTPError, URLError, Exception) as e:
            print(f"Attempt {i+1}/{MAX_RETRIES}: Broker not ready ({e})")
            time.sleep(RETRY_DELAY)
    print("Failed to connect to broker after max retries")
    return False


def parse_location(location_str: str) -> dict | None:
    """Parse location string to GeoJSON GeoProperty format."""
    if not location_str:
        return None
    
    # Handle "lat,lon" format (Point)
    if location_str.count(",") == 1 and "Polygon" not in location_str and "LineString" not in location_str:
        try:
            parts = location_str.replace('"', '').split(",")
            lat, lon = float(parts[0].strip()), float(parts[1].strip())
            return {
                "type": "GeoProperty",
                "value": {
                    "type": "Point",
                    "coordinates": [lon, lat]  # GeoJSON is [lon, lat]
                }
            }
        except (ValueError, IndexError):
            pass
    
    return None


def create_datetime_property(value: str) -> dict:
    """Create a DateTime Property in NGSI-LD format."""
    return {
        "type": "Property",
        "value": {
            "@type": "DateTime",
            "@value": value
        }
    }


def create_property(value: Any) -> dict:
    """Create a simple Property in NGSI-LD format."""
    return {
        "type": "Property",
        "value": value
    }


def parse_value(value: str) -> Any:
    """Parse string value to appropriate Python type."""
    if not value or value.strip() == "":
        return None
    
    value = value.strip()
    
    # Boolean
    if value.lower() == "true":
        return True
    if value.lower() == "false":
        return False
    
    # Numeric
    try:
        if "." in value:
            return float(value)
        return int(value)
    except ValueError:
        pass
    
    return value


def csv_row_to_ngsi_ld(row: dict, csv_type: str) -> dict | None:
    """Convert a CSV row to NGSI-LD entity in normalized format."""
    entity_id = row.get("id", "").strip()
    if not entity_id:
        return None
    
    # Get entity type and context for this CSV
    entity_type = TYPE_MAP.get(csv_type, csv_type)
    context_url = CONTEXT_MAP.get(csv_type, "https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld")
    attr_map = ATTRIBUTE_MAP.get(csv_type, {})
    
    entity = {
        "id": entity_id,
        "type": entity_type,
        "@context": [context_url]
    }
    
    # Handle observedAt -> dateObserved as DateTime Property
    if "observedAt" in row and row["observedAt"]:
        entity["dateObserved"] = create_datetime_property(row["observedAt"].strip())
    
    # Process each field
    for key, value in row.items():
        if key in ("id", "type", "observedAt", "@context") or not value:
            continue
        
        value = value.strip()
        if not value:
            continue
        
        # Map attribute name using the mapping table
        mapped_key = attr_map.get(key, key)
        
        # Handle location as GeoProperty
        if key == "location":
            geo_prop = parse_location(value)
            if geo_prop:
                entity["location"] = geo_prop
            continue
        
        # Parse and create Property
        parsed_value = parse_value(value)
        if parsed_value is not None:
            entity[mapped_key] = create_property(parsed_value)
    
    return entity


def upsert_entity(entity: dict) -> bool:
    """Upsert entity to broker (create or update)."""
    entity_id = entity["id"]
    context_url = entity.get("@context", [""])[0] if isinstance(entity.get("@context"), list) else entity.get("@context", "")
    
    headers = {
        "Content-Type": "application/ld+json",
        "Accept": "application/ld+json"
    }
    
    data = json.dumps(entity, ensure_ascii=False).encode("utf-8")
    
    # Try CREATE first
    try:
        req = Request(
            f"{BROKER_URL}/ngsi-ld/v1/entities",
            data=data,
            headers=headers,
            method="POST"
        )
        with urlopen(req, timeout=10) as resp:
            if resp.status in (200, 201):
                return True
    except HTTPError as e:
        if e.code == 409:  # Already exists, try UPDATE
            try:
                # Remove id, type for PATCH (keep @context)
                attrs = {k: v for k, v in entity.items() if k not in ("id", "type")}
                patch_data = json.dumps(attrs, ensure_ascii=False).encode("utf-8")
                
                req = Request(
                    f"{BROKER_URL}/ngsi-ld/v1/entities/{entity_id}/attrs",
                    data=patch_data,
                    headers=headers,
                    method="PATCH"
                )
                with urlopen(req, timeout=10) as resp:
                    return resp.status in (200, 204)
            except HTTPError as patch_e:
                print(f"  PATCH error for {entity_id}: {patch_e.code}")
                return False
        else:
            error_msg = e.read().decode() if hasattr(e, 'read') else str(e)
            print(f"  POST error for {entity_id}: {e.code} - {error_msg}")
            return False
    except URLError as e:
        print(f"  Network error for {entity_id}: {e}")
        return False
    
    return True


def load_csv_file(filepath: Path) -> int:
    """Load a CSV file and upsert all entities to broker."""
    print(f"\nProcessing {filepath.name}...")
    
    # Determine entity type from filename (remove .csv)
    csv_type = filepath.stem
    target_type = TYPE_MAP.get(csv_type, csv_type)
    context = CONTEXT_MAP.get(csv_type, "core")
    
    print(f"  Type: {csv_type} -> {target_type}")
    print(f"  Context: {context}")
    
    success_count = 0
    error_count = 0
    
    with open(filepath, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            entity = csv_row_to_ngsi_ld(row, csv_type)
            if entity:
                if upsert_entity(entity):
                    success_count += 1
                else:
                    error_count += 1
    
    print(f"  Result: {success_count} succeeded, {error_count} failed")
    return success_count


def main():
    """Main entry point."""
    print("=" * 60)
    print("HCMC Smart City Data Loader v2.0")
    print("=" * 60)
    print(f"Broker URL: {BROKER_URL}")
    print(f"Data Directory: {DATA_DIR}")
    
    if not wait_for_broker():
        sys.exit(1)
    
    data_path = Path(DATA_DIR)
    csv_files = sorted(data_path.glob("*.csv"))
    
    if not csv_files:
        print(f"No CSV files found in {DATA_DIR}")
        sys.exit(1)
    
    print(f"\nFound {len(csv_files)} CSV files:")
    for f in csv_files:
        print(f"  - {f.name}")
    
    total_success = 0
    for csv_file in csv_files:
        total_success += load_csv_file(csv_file)
    
    print("\n" + "=" * 60)
    print(f"Data loading complete! Total entities: {total_success}")
    print("=" * 60)


if __name__ == "__main__":
    main()
