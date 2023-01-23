export interface Station {
  Name: string;
  Longitude: number;
  Latitude: number;
  Code: number;
  Dist: number;
  Arrival: Map<number, Map<number, number>>
}
