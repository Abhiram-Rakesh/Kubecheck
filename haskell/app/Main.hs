{-# LANGUAGE OverloadedStrings #-}

module Main (main) where

import qualified Data.Aeson as A
import qualified Data.ByteString.Lazy as BSL
import System.Exit (exitFailure)
import System.IO (hPutStrLn, stderr)
import Validator

-- | Main entry point: read JSON resource from stdin, validate, output JSON violations
main :: IO ()
main = do
    input <- BSL.getContents
    case A.eitherDecode input of
        Left err -> do
            hPutStrLn stderr $ "Failed to parse JSON input: " ++ err
            exitFailure
        Right resource -> do
            let violations = validateResource resource
            BSL.putStr $ A.encode violations
