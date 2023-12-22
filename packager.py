import pandas as pd


def csv_to_parquet(csv_file_path, parquet_file_path):
    # Read the CSV file
    df = pd.read_csv(csv_file_path)

    # Write to a Parquet file
    df.to_parquet(parquet_file_path)


# Example usage
csv_file_path = "./output.csv"  # Replace with your CSV file path
parquet_file_path = "./output.parquet"  # Replace with your desired Parquet file path

csv_to_parquet(csv_file_path, parquet_file_path)
