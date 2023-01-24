export interface Station {
  Name: string;
  Longitude: number;
  Latitude: number;
  Code: number;
  Dist: number;
  // UTC + 0
  Arrival: { [dow: number]: { [hour: number]: number } };
}
